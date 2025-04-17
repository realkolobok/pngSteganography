# **STPNG - PNG Steganography Tool**  

(mini project made for fun)

A command-line tool for hiding and extracting messages in PNG images using LSB (Least Significant Bit) steganography with optional XOR encryption.  

## **Installation**  

### **Method 1: Run Directly (Requires Go)**  
```bash
go run stpng.go [flags]
```  

### **Method 2: Build and Run (Standalone Executable)**  
```bash
go build stpng.go  # Creates 'stpng' (or 'stpng.exe' on Windows)
./stpng [flags]    # Run the compiled binary
```  

## **How It Works**  

This tool hides messages in PNG images by modifying the least significant bits (LSBs) of pixel color channels (red, green, and blue). Each pixel can store 3 bits of data (one in each color channel), making the changes visually undetectable.  

### **Key Features**  
- **Message Encoding**: The message is converted to binary and embedded in the image pixels.  
- **Length Prefix**: A 4-byte header stores the message length for accurate extraction.  
- **XOR Encryption (Optional)**: Messages can be encrypted with a password using XOR cipher before embedding.  

## **Usage**  

### **1. Encode a Message into a PNG**  
```bash
./stpng -encode -i input.png -o output.png -m "secret" [-p password] [-e]
```  

**Flags:**  
- `-encode`: Enable encoding mode (required)  
- `-i`: Input PNG file (required)  
- `-o`: Output PNG file (default: `output.png`)  
- `-m`: Message to hide (required)  
- `-p`: Password for encryption (optional)  
- `-e`: Enable encryption (optional, use `-e=true` or just `-e`)  

**Example:**  
```bash
./stpng -encode -i cat.png -o secret.png -m "Hello" -p mypass -e
```  

### **2. Decode a Hidden Message from a PNG**  
```bash
./stpng -decode -i secret.png [-p password] [-e]
```  

**Flags:**  
- `-decode`: Enable decoding mode (required)  
- `-i`: Input PNG file (required)  
- `-p`: Password (required if encrypted)  
- `-e`: Enable decryption (optional, use `-e=true` or just `-e`)  

**Example:**  
```bash
./stpng -decode -i secret.png -p mypass -e
```  

## **Notes**  

- **Boolean Flags**: The `-e` flag does not require a value (just `-e` means `true`).  
- **Encryption**: If `-e` is used, a password (`-p`) is required.  

## **Limitations**  

- **Message Size**: Limited by the number of pixels in the image.  
- **Security**: XOR encryption is basic and not suitable for highly sensitive data.  

## **Help**  

For a full list of options, run:  
```bash
./stpng -help
```  

## **License**  

MIT License - Free for personal and commercial use.
