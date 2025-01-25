const std = @import("std");
const types = @import("types.zig");

pub fn subArrayUntil(data: []const u8, c: u8) []const u8 {
    var i: usize = 0;
    for (data) |value| {
        if (value == c) {
            return data[0..i];
        }
        i += 1;
    }
    return data;
}

const InputBuffer = struct {
    buffer: []u8,

    // Create a new input buffer
    pub fn new(allocator: *std.mem.Allocator, initial_size: usize) !*InputBuffer {
        // Allocate memory for the InputBuffer struct
        const input_buffer = try allocator.create(InputBuffer);

        // Allocate memory for the buffer
        input_buffer.buffer = try allocator.alloc(u8, initial_size);

        return input_buffer;
    }

    // Free the allocated memory
    pub fn free(self: *InputBuffer, allocator: *std.mem.Allocator) void {
        allocator.free(self.buffer);
        allocator.destroy(self);
    }

    // Read input from stdin into the buffer
    pub fn read_input(self: *InputBuffer, stdin: anytype) !void {
        @memset(self.buffer, 0);

        // Read input into the buffer
        const slice = try stdin.readUntilDelimiterOrEof(self.buffer, '\n');
        if (slice == null) {
            std.debug.print("Error reading input\n", .{});
            std.process.exit(1); // Exit with failure
        }

        const s = subArrayUntil(self.buffer, 0);
        self.buffer[s.len - 1] = 0;

        printInput(self, ".");
    }

    pub fn printInput(self: *InputBuffer, msg: []const u8) void {
        const s = subArrayUntil(self.buffer, 0);
        std.debug.print("{s} {} {s}\n", .{ msg, s.len, s });
    }

    // Handles meta commands like ".exit"
    pub fn doMetaCommand(self: *InputBuffer) types.MetaCommandResult {
        if (std.mem.startsWith(u8, self.buffer, ".exit")) {
            std.process.exit(0);
        } else {
            return types.MetaCommandResult.UnrecognizedCommand;
        }
    }

    // Prepares a statement (e.g., parses "insert" or "select")
    pub fn prepareStatement(self: *InputBuffer, statement: *types.Statement) types.PrepareResult {
        if (std.mem.startsWith(u8, self.buffer, "insert")) {
            statement.typ = types.StatementType.Insert;
            return types.PrepareResult.Success;
        } else if (std.mem.eql(u8, self.buffer, "select")) {
            statement.typ = types.StatementType.Select;
            return types.PrepareResult.Success;
        }

        return types.PrepareResult.UnrecognizedStatement;
    }

    // Executes a prepared statement
    pub fn executeStatement(statement: *types.Statement) void {
        switch (statement.typ) {
            types.StatementType.Insert => std.debug.print("This is where we would do an insert.\n", .{}),
            types.StatementType.Select => std.debug.print("This is where we would do a select.\n", .{}),
        }
    }
};

pub fn main() !void {
    var allocator = std.heap.page_allocator;
    const stdin = std.io.getStdIn().reader();

    // Create a new input buffer
    var input_buffer = try InputBuffer.new(&allocator, 128);
    defer input_buffer.free(&allocator);

    // Main loop
    while (true) {
        std.debug.print(">>> ", .{});

        // Read user input
        try input_buffer.read_input(&stdin);

        if (input_buffer.buffer[0] == '.') {
            switch (input_buffer.doMetaCommand()) {
                types.MetaCommandResult.Success => continue,
                types.MetaCommandResult.UnrecognizedCommand => {
                    input_buffer.printInput("Unrecognized command ");
                    continue;
                },
            }
        }

        var statement = types.Statement{ .typ = types.StatementType.Insert };
        // Switch for prepare_statement
        switch (input_buffer.prepareStatement(&statement)) {
            types.PrepareResult.Success => {},
            types.PrepareResult.UnrecognizedStatement => {
                input_buffer.printInput("Unrecognized keyword at start of ");
                continue;
            },
        }

        // Execute the statement
        InputBuffer.executeStatement(&statement);
        std.debug.print("Executed.\n", .{});
    }
}

test "simple test" {
    var list = std.ArrayList(i32).init(std.testing.allocator);
    defer list.deinit(); // try commenting this out and see if zig detects the memory leak!
    try list.append(42);
    try std.testing.expectEqual(@as(i32, 42), list.pop());
}
