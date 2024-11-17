// pkg/auto/pulumi.go - Functions for deploying and destroying Pulumi stacks
package auto

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/jaxxstorm/pedloy/pkg/graph"
	proj "github.com/jaxxstorm/pedloy/pkg/project"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/events"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func createOutputLogger(fields ...zap.Field) *zap.Logger {
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	core := zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), zapcore.DebugLevel)

	sampling := zapcore.NewSamplerWithOptions(
		core,
		time.Second,
		3,
		0,
	)

	// Add global fields to the logger
	return zap.New(sampling).With(fields...)
}

func processEvents(logger *zap.Logger, eventChannel <-chan events.EngineEvent) {
	for event := range eventChannel {
		jsonData, err := json.Marshal(event)
		if err != nil {
			logger.Error("Failed to marshal event to JSON", zap.Error(err))
			continue
		}
		logger.Info(string(jsonData))
	}
}

func createOrSelectStack(ctx context.Context, org string, stackName string, project proj.Project, source proj.ProjectSource) (auto.Stack, error) {
	var usedStackName string
	if org == "" {
		usedStackName = stackName
	} else {
		usedStackName = org + "/" + stackName
	}
	projectPath := filepath.Join(source.LocalPath, project.Name)
	return auto.UpsertStackLocalSource(ctx, usedStackName, projectPath)
}

func deployStack(project proj.Project, stack string, org string, source proj.ProjectSource, ctx context.Context, logger *zap.Logger, jsonLog bool) error {
	logger = logger.With(zap.String("project", project.Name), zap.String("stack", stack))
	logger.Info("Deploying stack")

	eventChannel := make(chan events.EngineEvent)
	go processEvents(logger, eventChannel)

	s, err := createOrSelectStack(ctx, org, stack, project, source)
	if err != nil {
		logger.Error("Failed to create or select stack", zap.Error(err))
		return err
	}

	var upErr error
	if jsonLog {
		_, upErr = s.Up(ctx, optup.EventStreams(eventChannel))
	} else {
		_, upErr = s.Up(ctx, optup.ProgressStreams(os.Stdout))
	}
	if upErr != nil {
		logger.Error("Failed to deploy stack", zap.Error(upErr))
	} else {
		logger.Info("Successfully deployed stack")
	}
	return upErr
}

func Deploy(org string, projects []proj.Project, source proj.ProjectSource, jsonLogger bool) {
	// Create a logger with a global field for deployment
	logger := createOutputLogger(zap.String("operation", "deploy"))
	defer logger.Sync()

	logger.Info("Starting deployment")

	// Get execution groups
	executionGroups, err := graph.GetExecutionGroups(projects)
	if err != nil {
		logger.Fatal("Failed to determine execution groups", zap.Error(err))
	}

	// Log the execution schedule
	logger.Info("Execution Schedule")
	for i, group := range executionGroups {
		logger.Info("Deployment Stage",
			zap.Int("stage", i+1),
			zap.Strings("deployments", group))
	}

	ctx := context.Background()
	deployed := make(map[string]bool)
	mu := &sync.Mutex{}

	// Execute each group sequentially
	for groupIndex, group := range executionGroups {
		stageLogger := logger.With(zap.Int("stage", groupIndex+1))
		stageLogger.Info("Executing deployment stage")

		var groupWG sync.WaitGroup
		groupErrors := make(chan error, len(group))

		// Deploy all items in the group concurrently
		for _, vertex := range group {
			groupWG.Add(1)
			go func(vertex string) {
				defer groupWG.Done()

				// Parse project and stack from vertex ID
				parts := strings.Split(vertex, ":")
				projectName, stackName := parts[0], parts[1]

				// Find the project definition
				var projectDef proj.Project
				for _, p := range projects {
					if p.Name == projectName {
						projectDef = p
						break
					}
				}

				// Deploy the stack
				stackLogger := stageLogger.With(
					zap.String("project", projectName),
					zap.String("stack", stackName),
				)
				stackLogger.Info("Deploying stack")
				err := deployStack(projectDef, stackName, org, source, ctx, stackLogger, jsonLogger)
				if err != nil {
					groupErrors <- fmt.Errorf("failed to deploy %s: %w", vertex, err)
					return
				}

				// Mark as deployed
				mu.Lock()
				deployed[vertex] = true
				mu.Unlock()
			}(vertex)
		}

		// Wait for all deployments in this group to complete
		groupWG.Wait()
		close(groupErrors)

		// Check for any errors in this group
		for err := range groupErrors {
			if err != nil {
				stageLogger.Fatal("Deployment failed", zap.Error(err))
			}
		}

		stageLogger.Info("Completed deployment stage")
	}

	logger.Info("Deployment completed successfully")
}

func Destroy(org string, projects []proj.Project, source proj.ProjectSource, jsonLogger bool) {
	// Create a logger with a global field for destruction
	logger := createOutputLogger(zap.String("operation", "destroy"))
	defer logger.Sync()

	logger.Info("Starting destruction")

	// Get execution groups
	executionGroups, err := graph.GetExecutionGroups(projects)
	if err != nil {
		logger.Fatal("Failed to determine execution groups", zap.Error(err))
	}

	// Log the destruction schedule in reverse order
	logger.Info("Destruction Schedule")
	for i := len(executionGroups) - 1; i >= 0; i-- {
		logger.Info("Destruction Stage",
			zap.Int("stage", len(executionGroups)-i),
			zap.Strings("stacks", executionGroups[i]))
	}

	ctx := context.Background()
	destroyed := make(map[string]bool)
	mu := &sync.Mutex{}

	// Execute each group sequentially in reverse order
	for i := len(executionGroups) - 1; i >= 0; i-- {
		group := executionGroups[i]
		stageLogger := logger.With(zap.Int("stage", len(executionGroups)-i))
		stageLogger.Info("Executing destruction stage")

		var groupWG sync.WaitGroup
		groupErrors := make(chan error, len(group))

		// Destroy all items in the group concurrently
		for _, vertex := range group {
			groupWG.Add(1)
			go func(vertex string) {
				defer groupWG.Done()

				// Parse project and stack from vertex ID
				parts := strings.Split(vertex, ":")
				projectName, stackName := parts[0], parts[1]

				// Find the project definition
				var projectDef proj.Project
				for _, p := range projects {
					if p.Name == projectName {
						projectDef = p
						break
					}
				}

				// Create event channel for this stack
				eventChannel := make(chan events.EngineEvent)
				go processEvents(stageLogger.With(
					zap.String("project", projectName),
					zap.String("stack", stackName),
				), eventChannel)

				// Create or select the stack
				s, err := createOrSelectStack(ctx, org, stackName, projectDef, source)
				if err != nil {
					groupErrors <- fmt.Errorf("failed to select stack %s: %w", vertex, err)
					return
				}

				// Destroy the stack
				var destroyErr error
				if jsonLogger {
					_, destroyErr = s.Destroy(ctx, optdestroy.EventStreams(eventChannel))
				} else {
					_, destroyErr = s.Destroy(ctx, optdestroy.ProgressStreams(os.Stdout))
				}

				if destroyErr != nil {
					groupErrors <- fmt.Errorf("failed to destroy %s: %w", vertex, destroyErr)
					return
				}

				// Mark as destroyed
				mu.Lock()
				destroyed[vertex] = true
				mu.Unlock()

				stageLogger.Info("Successfully destroyed stack",
					zap.String("project", projectName),
					zap.String("stack", stackName))
			}(vertex)
		}

		// Wait for all destructions in this group to complete
		groupWG.Wait()
		close(groupErrors)

		// Check for any errors in this group
		for err := range groupErrors {
			if err != nil {
				stageLogger.Fatal("Destruction failed", zap.Error(err))
			}
		}

		stageLogger.Info("Completed destruction stage")
	}

	logger.Info("Destruction completed successfully")
}
