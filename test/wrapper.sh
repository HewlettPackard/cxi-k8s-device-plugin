#!/bin/sh

go mod tidy
echo -e "\nRunning tests...\n"
go test ./test
if [ $? -ne 0 ]; then
    echo "Tests failed. Please check the output above."
    exit 1
else
    echo -e "\n\nAll tests passed successfully.\n\n"
fi

