#!/usr/bin/env ash

# Argument sent by the GO APP
# It's the folder to watch
FILES="$1"
cd $FILES
# In the folder to watch, each time a new file is created 
# We rename the file to replace the space by the dot to be ale to be read by the GO app 
for f in *; do mv "$f" `echo $f | tr ' ' '.'`; done
