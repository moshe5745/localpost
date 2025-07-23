# Localpost

Localpost is a CLI API client for storing, and executing HTTP request collections, with support for environment variables and dynamic response handling.

![Version](https://img.shields.io/github/v/release/moshe5745/localpost?logo=go&style=flat)
![Build Status](https://img.shields.io/github/actions/workflow/status/moshe5745/localpost/release.yml?logo=github&style=flat)
![License](https://img.shields.io/github/license/moshe5745/localpost?style=flat)

## Features

- **Auto Generation for Request Definition**
- **Cookies Handling**
- **Auto Login Before Requests**
- **Environment Variables**
- **Baseline testing**

## How it works?

#### Request definition example:
```yaml
headers:
  Accept: application/json
  Content-Type: application/json
body:
  json:
    username: user
    password: pass
set-env-var:
  TOKEN:
    body: jwt-token
```

You can easily create those definitions with: `lpost add-request`

Now you can execute this request with `lpost request POST_login` or with shorthand flag `lpost -r POST_login`.

> ℹ️ **Collaboration**: Commit your request definitions files to your repo to collaborate with others, or manage them locally without sharing. All the files is stored in `lpost`.

## Installation

- Grab the latest release from [GitHub Releases](https://github.com/moshe5745/localpost/releases)
  - ### Download the binary
    - #### macOS Intel (amd64)
      ```bash
      curl -L https://github.com/moshe5745/localpost/releases/latest/download/localpost_darwin_amd64.tar.gz -o lpost.tar.gz
      tar -xzf lpost.tar.gz
      ```
    - #### macOS Apple Silicon (arm64)
      ```bash
      curl -L https://github.com/moshe5745/localpost/releases/latest/download/localpost_darwin_arm64.tar.gz -o lpost.tar.gz
      tar -xzf lpost.tar.gz
      ```
    - #### Linux (amd64)
      ```bash
      curl -L https://github.com/moshe5745/localpost/releases/latest/download/localpost_linux_amd64.tar.gz -o lpost.tar.gz
      tar -xzf lpost.tar.gz
      ```
    - #### Linux (arm64)
      ```bash
      curl -L https://github.com/moshe5745/localpost/releases/latest/download/localpost_linux_arm64.tar.gz -o lpost.tar.gz
      tar -xzf lpost.tar.gz
      ```
    - #### Windows (x86_64)
      ```powershell
      curl -L https://github.com/moshe5745/localpost/releases/latest/download/localpost_windows_x86_64.zip -o lpost.zip
      Expand-Archive -Path lpost.zip -DestinationPath .
      ```
    - #### Windows (ARM64)
      ```powershell
      curl -L https://github.com/moshe5745/localpost/releases/latest/download/localpost_windows_arm64.zip -o lpost.zip
      Expand-Archive -Path lpost.zip -DestinationPath .
      ```
  - ### Make it executable
    ```bash
    chmod +x lpost
    ```
  - ### Move to /usr/local/bin (So it will be globally available in your machine)
    ```bash
    sudo mv lpost /usr/local/bin/lpost
    ```

### Shell Completion installation

- Enable autocompletion by adding the following to your shell config file.
  - Zsh
    ```zsh
    # Add to ~/.zshrc
    source <(lpost completion --shell zsh)
    ```
  - Bash
    ```bash
    # Add to ~/.bashrc
    source <(lpost completion --shell bash)
    ```
  - Fish
    ```bash
    # Add to ~/.config/fish/config.fish
    source (lpost completion --shell fish | psub)
    ```
  - PowerShell
    ```bash
    # Download automated script(setup-completion.ps1) from this repo and run it in your machine.
    # .\setup-completion.ps1

    ```
- Use TAB key for completion
  > ⚠️ After adding the completion line to your shell config (e.g., ~/.zshrc), run `source ~/.zshrc` (or equivalent) to apply it immediately, or restart your shell.

## Basic usage example

```bash
lpost add-request
# POST_login added
```

```bash
lpost set-env prod
# Default env is dev
```

```bash
lpost set-env-var BASE_URL https://example.com
```

```bash
lpost -r POST_login
$:
+--------+------+-------------------------+
| STATUS | TIME | BODY                    |
+--------+------+-------------------------+
| 200    | 29ms |     {                   |
|        |      |       "TOKEN": "123456" |
|        |      |     }                   |
+--------+------+-------------------------+
```

```bash
lpost -r POST_login -v
# Verbose for debugging
$:
-----
Status: 200 OK
Time: 18ms
URL: http://localhost:8080/login
-----
Request
  Headers:
    Content-Type: application/json
    ...
Request
  Body:

-----
Response
  Headers:
    Content-Type: [application/json]
    ...
Request
  Body:
    {"TOKEN":"123456"}
```

## Environment Variables

- Use `lpost` to store and manage variables for different environments (e.g., `dev`, `prod`).
- Let you set custom values (e.g., API endpoints, credentials) per environment.

> ℹ️ **Note**: These environment variables are unique to `localpost` and stored in `config.yaml`—they’re separate from your shell’s environment, `.env` files, or other tools.

### Setting Environment Variables

You can set variables in two ways:

- **CLI**:
  ```bash
  lpost set-env-var BASE_URL https://api.example.com
  ```
- **YAML**:

  ```yaml
  ---
  set-env-var:
    TOKEN: # Var name
      body: jwt-token # from "jwt-token" param in JSON body
    Cookie:
      header: Cookie # from "Cookie" header
  ```

  Then you can use them inside your requests definitions:

```yaml
url: "{BASE_URL}/login"
headers:
  Cookie: "{Cookie}"
```

## Configuration file

For storing the collection env vars and current env used.

- **File**: `config.yaml` in the `lpost` directory.
- **Format**:
  ```yaml
  env: dev # Current env used
  envs: # All envs list
    dev:
      BASE_URL: https://api.dev.com
      TOKEN: 123
    login:
      request: POST_login
      triggered_by: [400, 401, 403] #Default is [401]
    prod:
      BASE_URL: https://api.prod.com
      TOKEN: 456
  ```
  > ℹ️ Note: `config.yaml` is created automatically on first use with a default `env: dev` if it doesn’t exist.

### Request YAML Format

Request files are stored in the `requests/` directory. Example (`requests/POST_login.yaml`):

```yaml
url: "{BASE_URL}/login"
headers:
  Accept: application/json
  Content-Type: application/json
body:
  json:
    username: user
    password: pass
set-env-var:
  TOKEN:
    body: jwt-token
```

## Commands full list

| Command                     | Description                                                                                                      | Example Usage                                                            |
| --------------------------- | ---------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------ |
| `init`                      | Initialize `localpost`. creates `lpost/` directory with `config.yaml` and `requests/`.                           | `$: lpost init`                                                          |
| `add-request`               | Create a new request YAML file interactively with prompts for nickname, URL, method, and body type.              | `$: lpost add-request`                                                   |
| `request <METHOD_name>`     | Execute a request from a YAML file in `requests/`. Use `--infer-schema` to generate JTD schema. Shorthand: `-r`. | `$: lpost -r POST_login` or `$: lpost request GET_config --infer-schema` |
| `test`                      | Run all requests in `requests/` and validate responses against stored JTD schemas in `schemas/`.                 | `$: lpost test`                                                          |
| `set-env <env>`             | Set the current environment in `config.yaml`.                                                                    | `$: lpost set-env prod`                                                  |
| `set-env-var <key> <value>` | Set an environment variable for the current environment in `config.yaml`.                                        | `$: lpost set-env-var BASE_URL https://api.example.com`                  |
| `show-env`                  | Display the current environment and variables from `config.yaml`. Use `--all` for the full config.               | `$: lpost show-env` or `$: lpost show-env --all`                         |
| `completion`                | Output completion script for your shell (bash, zsh, fish) to stdout. Requires `--shell` flag.                    | `$: source <(lpost completion --shell zsh)`                              |
| `list`                | Lists all available requests | `$: source <(lpost list)`                              |

- **Global Flag**: Override the environment temporarily with `-e` or `--env`:

  ```bash
  lpost -e prod -r GET_users
  ```

- **Body Types**:
  - `json`: JSON object (e.g., `{"key": "value"}`).
  - `form-urlencoded`: Key-value pairs (e.g., `key=value`).
  - `form-data`: Fields and files (e.g., `fields: {field: value}`, `files: {file: path}`).
  - `text`: Plain text (e.g., `example text`).

## Building from Source

1. Clone the repo:
   ```bash
   git clone https://github.com/moshe5745/localpost.git
   cd localpost
   ```
2. Build:
   ```bash
   go build -o localpost
   ```
3. Run:
   ```bash
   lpost add-request
   ```

## Roadmap

Planned features to enhance `localpost`—contributions welcome!

### Next Release (v1.1.0)

- **Schema Snapshot Tests**: Compare response schemas against saved snapshots for regression testing (`lpost test`).

### Future Releases

- **OpenAPI Integration**: Import OpenAPI specs to auto-generate request files (`lpost import-openapi`).
- **Request Validation**: Define expected status codes or headers in YAML to validate responses.
- **Mock Server Mode**: Run `lpost` as a mock API server using request files (`lpost mock`).
- **Request Templates**: Reuse common request parts from template files.

> ℹ️ **Got ideas?**: Share them at [github.com/moshe5745/localpost](https://github.com/moshe5745/localpost)!

## License

[MIT License](LICENSE)

```

```
