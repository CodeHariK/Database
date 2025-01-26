CC = gcc
CFLAGS = -Wall -std=c11

SRC = src/main.c
TARGET = db

cbuild:
	$(CC) $(CFLAGS) $(SRC) -o $(TARGET)  && ./db mydb.db

clean:
	rm -f $(TARGET)

zigbuild:
	zig build run
