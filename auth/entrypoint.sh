wait-for "db:3306" -- "$@"

# Watch your .go files and invoke go build if the files changed.
CompileDaemon --build="go build -o main main.go"  --command=./main