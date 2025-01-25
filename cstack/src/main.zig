const std = @import("std");

const InputBuffer = struct {
    buffer: []u8, // A slice representing the buffer (similar to `char*` in C)
    buffer_length: usize, // Length of the buffer
    input_length: usize, // Length of the actual input

    // Create a new input buffer
    pub fn new(allocator: *std.mem.Allocator, initial_size: usize) !*InputBuffer {
        // Allocate memory for the InputBuffer struct
        const input_buffer = try allocator.create(InputBuffer);

        // Allocate memory for the buffer
        input_buffer.buffer = try allocator.alloc(u8, initial_size);
        input_buffer.buffer_length = initial_size;
        input_buffer.input_length = 0;

        return input_buffer;
    }

    // Free the allocated memory
    pub fn free(self: *InputBuffer, allocator: *std.mem.Allocator) void {
        allocator.free(self.buffer);
        allocator.destroy(self);
    }

    // Read input from stdin into the buffer
    pub fn read_input(self: *InputBuffer, stdin: anytype) !void {
        // Read input into the buffer
        const slice = try stdin.readUntilDelimiterOrEof(self.buffer, '\n');
        if (slice == null) {
            std.debug.print("Error reading input\n", .{});
            std.process.exit(1); // Exit with failure
        }

        const valid_slice = slice.?; // Unwrap the optional
        self.input_length = valid_slice.len;
        self.buffer[self.input_length] = 0; // Null-terminate the string
    }
};

pub fn main() !void {
    var allocator = std.heap.page_allocator;
    const stdin = std.io.getStdIn().reader();

    // Create a new input buffer
    var input_buffer = try InputBuffer.new(&allocator, 128);
    defer input_buffer.free(&allocator);

    // Print details about the input buffer
    std.debug.print(
        "Buffer: {s}, Buffer Length: {}, Input Length: {}\n",
        .{
            input_buffer.buffer,
            input_buffer.buffer_length,
            input_buffer.input_length,
        },
    );

    // Main loop
    while (true) {
        // Print the prompt
        std.debug.print(">>> ", .{});

        // Read user input
        try input_buffer.read_input(&stdin);

        // Convert input buffer to a null-terminated string for comparison
        const input_str = input_buffer.buffer[0..input_buffer.input_length];

        // Check if input is ".exit"
        if (std.mem.eql(u8, input_str, ".exit")) {
            std.debug.print("Exiting...\n", .{});
            break;
        } else {
            std.debug.print("Unrecognized command '{s}'.\n", .{input_str});
        }
    }
}

test "simple test" {
    var list = std.ArrayList(i32).init(std.testing.allocator);
    defer list.deinit(); // try commenting this out and see if zig detects the memory leak!
    try list.append(42);
    try std.testing.expectEqual(@as(i32, 42), list.pop());
}
