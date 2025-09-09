# githost-poc

A proof-of-concept for a simple Git server implemented in Go.

This application uses the Echo web framework to serve a bare Git repository over HTTP, allowing `git clone`, `git pull`, and `git push` operations.

## Building

To build the application, run:

```sh
go build .
```

## Running

To start the Git server, run the compiled binary:

```sh
./githost-poc
```

The server will start on port 8080.

## Usage

Once the server is running, you can interact with it using standard `git` commands.

### Cloning the repository

```sh
git clone http://localhost:8080/user/repo.git
```

### Pushing changes

1. Create a new commit in your local repository:
   ```sh
   cd repo
   echo "hello" > README.md
   git add README.md
   git commit -m "Initial commit"
   ```

2. Push the changes to the server:
   ```sh
   git push origin master
   ```
