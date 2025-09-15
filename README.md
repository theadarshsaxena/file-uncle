# File Uncle (WIP)

File Uncle (aka file-uncle) is a versatile file transfer tool used to receive files. Other features like sending files via http, file manager etc coming soon.

## Features

- **File Transfer**: Receive files from anywhere via http.
- **Serve Files**: Create a http file server to let anyone download the file from your system.

## Installation

To install File Uncle, follow these steps:

1. Clone the repository: `git clone https://github.com/theadarshsaxena/file-uncle.git`
2. Navigate to the project directory: `cd file-uncle`
3. Run the project: `go run .`

## Usage

1. Launch File Uncle by running `go run .` in your project directory or `go build .` followed by `./file-uncle`.
2. It will start a server at your chosen port (default: 8080). Clients in your network can send files from anywhere to your system.
3. (Optional) If you want, you can run ngrok to receive from outside of your local network.

### Sending Files via CLI (curl)

When using the `file-uncle receive` command, it starts the server at 8080 (by default) and it also provide a webUI to upload the files.
In addition to this, if you want to send the files via CLI in remote system, you the following commands:
To upload a file to the server using curl:

```sh
curl -F "uploadFile=@/path/to/your/file" http://localhost:8080/
```

If authentication is enabled:

```sh
curl -u username:password -F "uploadFile=@/path/to/your/file" http://localhost:8080/
```

### Encrypting and Decrypting Files

For better security when sending the file via tunnel, e.g., ngrok, prefer sending the encrypted file.
You can encrypt a file before sending and decrypt it after receiving using OpenSSL (AES-256) as follows:

**Encrypt before sending:**

```sh
openssl enc -aes-256-cbc -salt -in /path/to/your/file -out /path/to/your/file.enc -k yourpassword
```

**Send the encrypted file:**

```sh
curl -F "uploadFile=@/path/to/your/file.enc" http://localhost:8080/
```

**Decrypt after receiving:**

```sh
openssl enc -d -aes-256-cbc -in /path/to/received/file.enc -out /path/to/decrypted/file -k yourpassword
```

Replace `/path/to/your/file` and `yourpassword` with your actual file path and password.

## Contributing

Contributions are welcome! If you have any ideas, suggestions, or bug reports, please open an issue or submit a pull request. Make sure to follow our [contribution guidelines](CONTRIBUTING.md).

## License

This project is licensed under the [MIT License](LICENSE).

