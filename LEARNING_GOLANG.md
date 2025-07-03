# Learning Golang

<!-- TOC -->

- [Learning Golang](#learning-golang)
  - [Tutorials and documentation](#tutorials-and-documentation)
  - [Challenges](#challenges)
  - [Creating a package and module in first time](#creating-a-package-and-module-in-first-time)

<!-- TOC -->

## Tutorials and documentation

- https://blog.aeciopires.com/primeiros-passos-com-go/
- https://github.com/dariubs/GoBooks
- https://github.com/aeciopires/go_mysamples
- https://gitlab.com/aeciopires/kube-pires
- https://www.freecodecamp.org/news/learn-golang-handbook/
- https://www.codecademy.com/catalog/subject/devops
- https://www.codecademy.com/catalog/language/go
- http://aprendago.com
- https://golang.org/doc
- https://golangr.com
- https://go.dev
- https://pkg.go.dev
- https://learn.go.dev
- https://www.geeksforgeeks.org/golang/?ref=lbp
- https://astaxie.gitbooks.io/build-web-application-with-golang/content/pt-br/
- https://medium.com/@denis_santos/primeiros-passos-com-golang-c3368f6d707a
- https://sensedia.com/conteudo/introducao-ao-golang/
- https://sensedia.com/conteudo/motivos-para-utilizar-go
- https://sensedia.com/conteudo/estrutura-da-linguagem-go
- https://www.linode.com/docs/development/go/install-go-on-ubuntu/
- https://tecadmin.net/install-go-on-macos/
- https://sourabhbajaj.com/mac-setup/Go/README.html
- https://tour.golang.org
- https://www.youtube.com/watch?v=YS4e4q9oBaU
- https://www.youtube.com/watch?v=Q0sKAMal4WQ
- https://www.youtube.com/watch?v=G3PvTWRIhZA
- https://www.youtube.com/watch?v=_MkQLDMak-4
- https://www.youtube.com/watch?v=JepHr8egvBI
- https://stackify.com/learn-go-tutorials
- https://golangbot.com
- https://www.tutorialspoint.com/go/index.htm
- https://www.guru99.com/google-go-tutorial.html
- http://www.golangbr.org/doc/codigo
- https://golang.org/doc/articles/wiki
- https://gobyexample.com
- https://hackr.io/tutorials/learn-golang
- https://hackernoon.com/basics-of-golang-for-beginners-6bd9b40d79ae
- https://medium.com/hackr-io/learn-golang-best-go-tutorials-for-beginners-deb6cab45867
- https://github.com/dariubs/GoBooks#readme
- https://www.digitalocean.com/community/books/how-to-code-in-go-ebook-pt
- https://dzone.com/articles/golang-tutorial-learn-golang-by-examples
- https://dzone.com/articles/structure-of-a-go-program 
- https://gobyexample.com/
- https://gowebexamples.com
- https://mholt.github.io/json-to-go/
- https://awesome-go.com/
- https://research.swtch.com/
- https://levelup.gitconnected.com/get-a-taste-of-concurrency-in-go-625e4301810f
- https://golang.org/doc/effective_go.html
- https://www.golangprograms.com/go-language.html
- https://golang.org/ref/spec
- https://golang.org/pkg/fmt/

## Challenges

- https://exercism.org/tracks/go
- https://www.codewars.com/kata/search/go
- https://www.hackerrank.com/
- https://github.com/RajaSrinivasan/assignments
- https://github.com/cblte/100-golang-exercises
- https://gophercises.com/
- https://codingchallenges.fyi/blog/learn-go/
- https://labex.io/courses/go-practice-challenges
- https://www.codechef.com/practice/go
- https://medium.com/@prasgema/learn-golang-over-weekend-challenge-292fa89f80ca
- https://coderbyte.com/challenges?utm_campaign=NewHomepage
- https://gobyexample.com/

## Creating a package and module in first time

Create the package and folder structure.

```bash
cd pires-cli/app

# Hosting URL and module name
# The go.mod file with the dependency list will be created
go mod init github.com/aeciopires/pires-cli

# Get the tools to manage arguments and environment variables
go mod download

# Create the package structure and main CLI file
mkdir -p internal/config
mkdir -p internal/getinfo

# Initialize cobra-cli. Will be created app/cmd subdiretory and app/cmd/root.go file and app/main.go file
$HOME/go/bin/cobra-cli init --viper
# Create command and arguments. Will be created app/cmd/create-yaml.go file
$HOME/go/bin/cobra-cli add create-yaml --viper
```

References:
- Instructions on how to create a module and manage dependencies: https://eternaldev.com/blog/adding-and-removing-dependency-in-go-modules
- Best practices for documenting a package: https://go.dev/blog/godoc
- https://github.com/spf13/cobra
- https://github.com/spf13/cobra-cli/blob/main/README.md
- https://github.com/spf13/cobra/blob/main/site/content/user_guide.md
- https://labex.io/tutorials/go-how-to-manage-multiple-cli-subcommands-422495
- https://github.com/go-playground/validator
- https://pkg.go.dev/github.com/go-playground/validator/v10
- https://dev.to/kittipat1413/a-guide-to-input-validation-in-go-with-validator-v10-56bp
- https://www.highlight.io/blog/5-best-logging-libraries-for-go
- https://github.com/rs/zerolog
- https://github.com/go-git/go-git
