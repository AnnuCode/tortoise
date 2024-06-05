# Tortoise

An interactive TUI app to manage tasks. It renders tasks according to the approaching deadline, with higher priority tasks displayed first.

## Usage

1. Run Dragonfly: 

    - Linux
    ```
    docker run --network=host --ulimit memlock=-1 docker.dragonflydb.io/dragonflydb/dragonfly
    ```
    - MacOS
    ```
    docker run -p 6379:6379 --ulimit memlock=-1 docker.dragonflydb.io/dragonflydb/dragonfly
    ```

2. Run the App: 
```go
go run .
```
3. Use the arrow keys to move around the task lists. 
4. Press `n` to add a new task or `e` to edit a task.
5. Press `enter` key while focused on a task to update its status.
6. Press `q` or `ctrl+c` to quit the app, and restart it(`go run .`) to see the tasks rendered according to their deadlines.

## Todo

- add the feature to persist data to the disk
- display task deadlines in the terminal view