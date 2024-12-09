#!/bin/bash

# Define target platforms and architectures
targets=(
  "android/arm64"
  "darwin/amd64"
  "darwin/arm64"
  "linux/amd64"
  "linux/arm64"
  "windows/amd64"
  "windows/arm64"
)

output_dir="build"

# Ensure the output directory exists
mkdir -p $output_dir

# Build for each target
for target in "${targets[@]}"; do
  os=$(echo $target | cut -d '/' -f 1)
  arch=$(echo $target | cut -d '/' -f 2)
  output_name=$output_dir/ziba-$os-$arch

  # Append .exe for Windows binaries
  if [ "$os" == "windows" ]; then
    output_name+='.exe'
  fi

  echo "Building for $os/$arch..."
  GOOS=$os GOARCH=$arch go build -o $output_name

  if [ $? -ne 0 ]; then
    echo "Failed to build for $os/$arch"
    exit 1
  fi
done

echo "Build completed. Binaries are in the '$output_dir' directory."
