# Pedloy: Pulumi Environment Deployer

Pedloy is a CLI tool for managing Pulumi stacks, enabling the deployment and destruction of infrastructure environments in a defined order. It supports both local and Git-based project sources, offering concurrency and dependency resolution for efficient stack management.

## Features

- **Deployment and Destruction**: Deploy or destroy stacks with dependencies resolved in order.
- **Preview Plans**: View the order of operations before deploying or destroying.
- **Git Integration**: Use Git repositories for Pulumi project sources.
- **Customizable Configuration**: Define projects and dependencies in a YAML configuration file.
- **Logging**: Supports JSON logging for structured output.

## Installation

Clone the repository and build the binary:

```bash
git clone https://github.com/jaxxstorm/pedloy.git
cd pedloy
go build -o pedloy cmd/pedloy/main.go
```

Alternatively, you can use `go install`:

```bash
go install github.com/jaxxstorm/pedloy/cmd/pedloy@latest
```

## Usage

### Commands

- `deploy`: Deploy the stacks defined in your configuration.
- `destroy`: Destroy the stacks defined in your configuration.

### Flags

| Flag             | Description                                   | Default       |
|------------------|-----------------------------------------------|---------------|
| `--config`       | Path to the configuration file               | `projects.yml`|
| `--org`          | The Pulumi organization                      |               |
| `--path`         | Path to local Pulumi projects                |               |
| `--git-url`      | Git repository URL for Pulumi projects       |               |
| `--git-branch`   | Git branch to use                            | `main`        |
| `--preview`      | Preview the deployment or destruction plan   | `false`       |
| `--json`         | Enable JSON logging                          | `false`       |

### Examples

#### Deploying Stacks

```bash
pedloy deploy --config projects.yml --org my-org --path ./pulumi-projects
```

#### Destroying Stacks

```bash
pedloy destroy --config projects.yml --org my-org
```

#### Preview Deployment Plan

```bash
pedloy deploy --preview --config projects.yml
```

## Configuration

The configuration is defined in a YAML file. Here’s an example `projects.yml`:

```yaml
projects:
  - name: project-a
    stacks:
      - dev
      - prod
    dependsOn:
      - project-b
  - name: project-b
    stacks:
      - dev
      - prod
```

### Structure

- `name`: The name of the Pulumi project.
- `stacks`: A list of stacks for the project.
- `dependsOn`: Other projects this project depends on.

## Project Structure

```
pedloy/
├── cmd/
│   └── pedloy/
│       ├── deploy/
│       │   └── cli.go
│       ├── destroy/
│       │   └── cli.go
│       └── main.go
├── pkg/
│   ├── auto/
│   │   └── pulumi.go
│   ├── config/
│   │   └── load.go
│   ├── graph/
│   │   └── graph.go
│   ├── project/
│   │   └── projects.go
│   └── utils/
│       ├── preview.go
│       └── validation.go
├── README.md
```

## Development

### Prerequisites

- [Go](https://golang.org/doc/install)
- [Pulumi CLI](https://www.pulumi.com/docs/get-started/install/)

### Building the Binary

```bash
go build -o pedloy cmd/pedloy/main.go
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

Enjoy deploying your infrastructure with **Pedloy**!