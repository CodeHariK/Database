const std = @import("std");
const types = @import("types.zig");

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

        const s = types.subArrayUntil(self.buffer, 0);
        self.buffer[s.len - 1] = 0;

        printInput(self, ".");
    }

    pub fn printInput(self: *InputBuffer, msg: []const u8) void {
        const s = types.subArrayUntil(self.buffer, 0);
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
            statement.type = types.StatementType.Insert;

            // // Using std.fmt.Scanner to extract values
            // var scanner = std.fmt.Scanner{
            //     .input = self.buffer,
            // };

            // const id = try scanner.readInt(u32);
            // const username = try scanner.readUntilDelimiterOrEof();
            // const email = try scanner.readUntilDelimiterOrEof();

            // // Copy parsed values into the statement
            // std.mem.copy(u8, &statement.row_to_insert.username, username);
            // std.mem.copy(u8, &statement.row_to_insert.email, email);
            // statement.row_to_insert.id = id;

            return types.PrepareResult.Success;
        } else if (std.mem.eql(u8, self.buffer, "select")) {
            statement.type = types.StatementType.Select;
            return types.PrepareResult.Success;
        }

        return types.PrepareResult.UnrecognizedStatement;
    }
};

pub fn main() !void {
    var allocator = std.heap.page_allocator;
    const stdin = std.io.getStdIn().reader();

    var table = try types.Table.new(&allocator);

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

        var statement: types.Statement = undefined;
        switch (input_buffer.prepareStatement(&statement)) {
            types.PrepareResult.Success => {},
            types.PrepareResult.SyntaxError => std.debug.print("Syntax error. Could not parse statement.\n", .{}),
            types.PrepareResult.UnrecognizedStatement => {
                input_buffer.printInput("Unrecognized keyword at start of ");
                continue;
            },
        }

        // Execute the statement
        switch (table.executeStatement(&statement)) {
            types.ExecuteResult.Success => {},
            types.ExecuteResult.TableFull => std.debug.print("Error: Table full.\n", .{}),
        }
        std.debug.print("Executed.\n", .{});
    }
}

test "simple test" {
    var list = std.ArrayList(i32).init(std.testing.allocator);
    defer list.deinit(); // try commenting this out and see if zig detects the memory leak!
    try list.append(42);
    try std.testing.expectEqual(@as(i32, 42), list.pop());
}
