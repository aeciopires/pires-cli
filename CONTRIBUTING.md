<!-- TOC -->

- [Requirements](#requirements)
- [Learning Golang](#learning-golang)
- [Contributing](#contributing)
- [About Visual Code (VSCode)](#about-visual-code-vscode)

<!-- TOC -->

# Requirements

- Configure authentication on your Github account to use the SSH protocol instead of HTTP. Watch this tutorial to learn how to set up: https://help.github.com/en/github/authenticating-to-github/adding-a-new-ssh-key-to-your-github-account
- Install the commands: ``asdf``, ``golang``, ``make``, ``gcloud``, ``kubectl``, ``docker``, ``git``, ``helm-docs``, ``trivy``, helm plugins and other softwares following the tutorials:
  - [MacOS](https://github.com/aeciopires/adsoft/blob/master/softwares-macos.md)
  - [Ubuntu](https://github.com/aeciopires/adsoft/blob/master/softwares-ubuntu.md)
- Login in GCP using gcloud:

```bash
gcloud init

# The default browser will open to complete login and grant permissions.
gcloud auth login
gcloud auth application-default login

# Authentication in GCP-Artifact Registry
gcloud auth configure-docker gcr.io

# See configurations of gcloud
gcloud config list
```

# Learning Golang

See [LEARNING_GOLANG.md](LEARNING_GOLANG.md) file.

# Contributing

- Create a fork this repository.
- Clone the forked repository to your local system:

```bash
git clone FORKED_REPOSITORY
```

- Add the address for the remote original repository:

```bash
git remote -v
git remote add upstream git@github.com:aeciopires/pires-cli.git
git remote -v
```

- Configure your name and email used during commits:

> Attention!!! Change your name and surname in commands bellow

```bash
cd pires-cli

git config --local user.name "Your_name Your_surname"

git config --local user.email "your_email"
```

- Install dependencies.

```bash
cd pires-cli/app
make prepare
```

- Create a branch. Example:

```bash
git checkout -b BRANCH_NAME
```

> - **TIP!!!** Try use conventionnal names for branch and commits explained in follow pages:
>   - https://medium.com/@shinjithkanhangad/git-good-best-practices-for-branch-naming-and-commit-messages-a903b9f08d68
>   - https://www.conventionalcommits.org/en/v1.0.0/
>   - https://dev.to/varbsan/a-simplified-convention-for-naming-branches-and-commits-in-git-il4
>   - https://gist.github.com/qoomon/5dfcdf8eec66a051ecd85625518cfd13

- Make sure you are on the correct branch using the following command. The branch in use contains the '*' before the name.

```bash
git branch
```

- Make your changes and tests.

> Attention!!!
>
> 1. Change the ``app/CHANGELOG.md`` when you make changes.
> 2. Dependencies version like ``yq`` needs changes in ``app/Makefile`` ``app/Dockerfile`` files
> 3. Change the value of ``CLIVersion`` variable in ``app/internal/config/config.go`` when you make changes to your Golang application. Follow the Semantic Version (https://semver.org/).
> 4. Change image Docker version in ``app/docker-compose.yml`` file

- Commit the changes to the branch.
- Push files to repository remote with command:

```bash
git push --set-upstream origin BRANCH_NAME
```

- Create Pull Request (PR) to the `main` branch. See this [tutorial](https://help.github.com/en/github/collaborating-with-issues-and-pull-requests/creating-a-pull-request-from-a-fork)
- Update the content with the suggestions of the reviewer (if necessary).
- After your pull request is merged to the `main` branch, update your local clone:

```bash
git checkout main
git pull upstream main
```

- Clean up after your pull request is merged with command:

```bash
git branch -d BRANCH_NAME
```

- Then you can update the ``main`` branch in your forked repository.

```bash
git push origin main
```

- And push the deletion of the feature branch to your GitHub repository with command:

```bash
git push --delete origin BRANCH_NAME
```

- Create a new tag to generate a new release of the package using the following commands:

```bash
cd pires-cli/app
export CLI_VERSION="$(go run . -v)"
echo "$CLI_VERSION"
git tag -a "$CLI_VERSION" -m "New release"
git push --tags
```

- Build the CLI using these commands:

```bash
cd pires-cli/app
make build
```

Create a release in Github and add the artifacts located in ``app/bin`` directory.

- To keep your fork in sync with the original repository, use these commands:

```bash
git pull upstream main
git pull upstream main --tags

git push origin main
git push origin main --tags
```

References:

- https://blog.scottlowe.org/2015/01/27/using-fork-branch-git-workflow/

# About Visual Code (VSCode)

Use a IDE (Integrated Development Environment) or text editor of your choice. By default, the use of VSCode is recommended.

VSCode (https://code.visualstudio.com), combined with the following plugins, helps the editing/review process, mainly allowing the preview of the content before the commit, analyzing the Markdown syntax and generating the automatic summary, as the section titles are created/changed.

Plugins to Visual Code:

- docker: https://marketplace.visualstudio.com/items?itemName=ms-azuretools.vscode-docker (require docker-ce package)
- gitlens: https://marketplace.visualstudio.com/items?itemName=eamodio.gitlens (require git package)
- golang: https://marketplace.visualstudio.com/items?itemName=golang.go
- gotemplate-syntax: https://marketplace.visualstudio.com/items?itemName=casualjim.gotemplate
- Markdown-all-in-one: https://marketplace.visualstudio.com/items?itemName=yzhang.markdown-all-in-one
- Markdown-lint: https://marketplace.visualstudio.com/items?itemName=DavidAnson.vscode-markdownlint
- Markdown-toc: https://marketplace.visualstudio.com/items?itemName=AlanWalk.markdown-toc
- YAML: https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml
- Helm Intellisense: https://marketplace.visualstudio.com/items?itemName=Tim-Koehler.helm-intellisense

Theme for VSCode:

- https://code.visualstudio.com/docs/getstarted/themes
- https://dev.to/thegeoffstevens/50-vs-code-themes-for-2020-45cc
- https://vscodethemes.com/
