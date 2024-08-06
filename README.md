# File Uncle (WIP)

File Uncle (aka file-uncle) is a versatile file transfer tool used to receive files. Other features like sending files via http, file manager etc coming soon.

## Features

- **File Transfer**: Receive files from anywhere via http.

## Installation

To install File Uncle, follow these steps:

1. Clone the repository: `git clone https://github.com/theadarshsaxena/file-uncle.git`
2. Navigate to the project directory: `cd file-uncle`
3. Run the project: `go run .`

## Usage

1. Launch File Uncle by running `go run .` in your project directory or `go build .` followed by `./file-uncle`.
2. It will start a server at port 8085, then, clients in your network can send file from anywhere to your system.
3. (Optional) If you want, you can run ngrok to receive from outside of your local network.

## Contributing

Contributions are welcome! If you have any ideas, suggestions, or bug reports, please open an issue or submit a pull request. Make sure to follow our [contribution guidelines](CONTRIBUTING.md).

## License

This project is licensed under the [MIT License](LICENSE).

