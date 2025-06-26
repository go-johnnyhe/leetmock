#!/bin/bash

echo "=== Testing Vim Autoread Fix ==="
echo "1. Open test.txt in vim in another terminal"
echo "2. Press ENTER when ready..."
read

echo "Writing change 1..."
echo "Change 1: $(date)" >> test.txt
sleep 2

echo "Writing change 2..."
echo "Change 2: $(date)" >> test.txt
sleep 2

echo "Writing change 3..."
echo "Change 3: $(date)" >> test.txt
sleep 2

echo "Done! Check if vim updated automatically without you switching windows."