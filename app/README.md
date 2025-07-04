# pires-cli

<!-- TOC -->

- [pires-cli](#pires-cli)
- [About](#about)
- [Requirements to use pires-cli](#requirements-to-use-pires-cli)
  - [Permissions](#permissions)
- [Software dependencies](#software-dependencies)
- [Run CLI](#run-cli)
  - [Downloading CLI](#downloading-cli)
- [How to use](#how-to-use)
  - [STEP-0: Login/Logout on GCP](#step-0-loginlogout-on-gcp)
  - [STEP-1: Getting version and help about the pires-cli](#step-1-getting-version-and-help-about-the-pires-cli)
    - [Enable debug mode](#enable-debug-mode)
  - [STEP-2: Create the configuration file before run the pires-cli](#step-2-create-the-configuration-file-before-run-the-pires-cli)
    - [Configuration file content or environment variables supported](#configuration-file-content-or-environment-variables-supported)
  - [GCP Actions](#gcp-actions)
    - [(OPTIONAL) Create service account](#optional-create-service-account)
    - [(OPTIONAL) Grant role to service account](#optional-grant-role-to-service-account)
    - [(OPTIONAL) Create database in GCP-CloudSQL (PostgreSQL)](#optional-create-database-in-gcp-cloudsql-postgresql)
    - [(OPTIONAL) Create database user in GCP-CloudSQL (PostgreSQL)](#optional-create-database-user-in-gcp-cloudsql-postgresql)
    - [(OPTIONAL) Export firewall rules to CSV file](#optional-export-firewall-rules-to-csv-file)
    - [(OPTIONAL) Export to TXT file the PostgreSQL audit logs (INSERT, UPDATE, DELETE) from a Cloud SQL instance](#optional-export-to-txt-file-the-postgresql-audit-logs-insert-update-delete-from-a-cloud-sql-instance)
    - [(OPTIONAL) Export to TXT file the PostgreSQL users and permissions from a Cloud SQL instance](#optional-export-to-txt-file-the-postgresql-users-and-permissions-from-a-cloud-sql-instance)

<!-- TOC -->

# About

``pires-cli`` is my CLI, developed in Golang, to perform Ops activities

See [README.md#how-to-use](README.md#how-to-use) section to more informations.

# Requirements to use pires-cli

## Permissions
> ATTENTION!!! You need to meet these requirements:
> - **GCP**: You need to have ``roles/owner`` associated with your user in each project of each environment;

# Software dependencies

Install all software dependencies following the instructions of the file [CONTRIBUTING.md#requirements](../CONTRIBUTING.md#requirements).

# Run CLI

## Downloading CLI

Get the latest version of ``pires-cli`` from https://github.com/aeciopires/pires-cli/releases according your operating system and architecture.
Save the binary in ``$HOME/pires-cli/`` directory (create it if necessary) and add permission to execute.

See [README.md#how-to-use](README.md#how-to-use) section to more informations.

# How to use

## STEP-0: Login/Logout on GCP

References:

- https://onecompiler.com/questions/3svmubqa9/how-to-logout-from-an-account-on-gcloud-sdk
- https://cloud.google.com/sdk/gcloud/reference/auth

Changes the values of variables bellow accordig your needs and run the commands:

```bash
# Configure gcloud. Uncomment only if you needs
#gcloud init

# The default browser will open to complete login and grant permissions.
gcloud auth login
gcloud auth application-default login

# See configurations of gcloud
gcloud config list

# Logout gcloud. Uncomment only if you needs
#gcloud auth revoke
#gcloud auth revoke --all
```

## STEP-1: Getting version and help about the pires-cli

```bash
$HOME/pires-cli/pires-cli -v # show short version
$HOME/pires-cli/pires-cli -V # Show long version, with architeture and operating system
$HOME/pires-cli/pires-cli -h # show global help

$HOME/pires-cli/pires-cli gcp -h # show help about gcp command

$HOME/pires-cli/pires-cli gcp cloudsql -h                 # show help about cloudsql command
$HOME/pires-cli/pires-cli gcp cloudsql create-user -h     # show help about create-user command
$HOME/pires-cli/pires-cli gcp cloudsql create-database -h # show help about create-database command

$HOME/pires-cli/pires-cli gcp iam -h             # show help about iam command
$HOME/pires-cli/pires-cli gcp iam create-role -h # show help about create-role command
$HOME/pires-cli/pires-cli gcp iam create-sa -h   # show help about create-sa command

$HOME/pires-cli/pires-cli gcp firewall -h              # show help about firewall command
$HOME/pires-cli/pires-cli gcp firewall export-rules -h # show help about export-rules command
```

### Enable debug mode

Enable debug mode using the ``-D`` for ``pires-cli`` in any position.

## STEP-2: Create the configuration file before run the pires-cli

> Attention!!! Order of precedence:
>
> 1) Configuration files have priority over environment variables and CLI options.
>
> 2) If no custom path with customization file is passed, the ``app/.env`` or ``/app/.env`` file will be considered and will have priority over CLI options.
>
> 3) If the ``app/.env`` or ``/app/.env`` file does not exist, environment variables (starting with ``CLI_``) will be given priority over CLI options.
>
> 4) If environment variables (starting with ``CLI_``) do not exist, CLI options will be considered.
>
> 5) If no CLI options are passed and there is no error message related to this, the default values ​​of ``pires-cli`` defined in the ``app/internal/config/config.go`` file will be considered.

Example:

```bash
mkdir -p $HOME/pires-cli/
touch $HOME/pires-cli/.env
```

Run the ``pires-cli`` command passing the new configuration file:

```bash
$HOME/pires-cli/pires-cli -C $HOME/pires-cli/.env
```

### Configuration file content or environment variables supported

The supported environment variables starting with ``CLI_`` and are defined in the ``app/internal/config/config.go`` file.

The environment variables is formed by ``CLI_`` prefix plus long flag name. Example: The value passed by ``--gcp-region`` can be passed by ``CLI_GCP_REGION``.

Run ``$HOME/pires-cli/pires-cli -h`` or ``pires-cli subcommand -h`` to see all long flag name and default values.

> Attention!!! The order is important because some variables is readed first and used to compose other variables.

```env
CLI_CONFIG_FILE=    # Dir of configuration file. Can be ommited. In this case, ``pires-cli`` follow the precedence rules explained in [README.md#configuration-file](README.md#configuration-file) section.
CLI_GCP_REGION=     # GCP region. Supported values in lower case. Example: us-central1
CLI_GCP_PROJECT=    # GCP project. Supported values in lower case. Example: nonprod
CLI_ENVIRONMENT=    # Environment name. Supported values in lower case: dev, staging and production. Example: dev
CLI_DATABASE_TYPE=  # Database type. Supported values in lower case: postgresql, mongodb and none. Example: postgresql
```

Other variables and values is formed during the execution.

## GCP Actions

### (OPTIONAL) Create service account

Create service account for application in specific project and environment.

```bash
$HOME/pires-cli/pires-cli gcp iam create-sa -C $HOME/pires-cli/.env -D -s kube-pires-gsa
```

### (OPTIONAL) Grant role to service account

Grant role to service account in specific project and environment.

```bash
$HOME/pires-cli/pires-cli gcp iam grant-role -C $HOME/pires-cli/.env -D -m "serviceAccount:kube-pires-gsa@nonprod.iam.gserviceaccount.com" -r "roles/cloudsql.editor"
```

### (OPTIONAL) Create database in GCP-CloudSQL (PostgreSQL)

Create database for application in specific project and environment.

```bash
$HOME/pires-cli/pires-cli gcp cloudsql create-database -C $HOME/pires-cli/.env -D -i nonprod-psql -d kube-pires-db
```

### (OPTIONAL) Create database user in GCP-CloudSQL (PostgreSQL)

Create database user for application in specific project and environment.

```bash
$HOME/pires-cli/pires-cli gcp cloudsql create-user -C $HOME/pires-cli/.env -D -i nonprod-psql -u kube-pires -p changeme
```

### (OPTIONAL) Export firewall rules to CSV file

Export firewall rules to CSV file in specific project and environment.

```bash
$HOME/pires-cli/pires-cli gcp firewall export-rules -C $HOME/pires-cli/.env -D -o $HOME
```

### (OPTIONAL) Export to TXT file the PostgreSQL audit logs (INSERT, UPDATE, DELETE) from a Cloud SQL instance

Export to TXT file the PostgreSQL audit logs (INSERT, UPDATE, DELETE) from a Cloud SQL instance

> ATTENTION!!!
> This requires the ``cloudsql.pgaudit`` flag to be enabled on the instance.

```bash
$HOME/pires-cli/pires-cli gcp cloudsql export-postgresql-audit-logs -i nonprod-psql -C $HOME/pires-cli/.env -D -o $HOME
```

### (OPTIONAL) Export to TXT file the PostgreSQL users and permissions from a Cloud SQL instance

Export to TXT file the PostgreSQL users and permissions from a Cloud SQL instance in specific project.

```bash
$HOME/pires-cli/pires-cli gcp cloudsql export-postgresql-users-permissions -i nonprod-psql -u postgres -C $HOME/pires-cli/.env -D -o $HOME
```
