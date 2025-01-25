const std = @import("std");

pub const ExecuteResult = enum {
    Success,
    TableFull,
};

pub const MetaCommandResult = enum {
    Success,
    UnrecognizedCommand,
};

pub const PrepareResult = enum {
    Success,
    SyntaxError,
    UnrecognizedStatement,
};

pub const StatementType = enum {
    Insert,
    Select,
};

pub const Statement = struct {
    type: StatementType,
    row_to_insert: Row,
};

const COLUMN_USERNAME_SIZE = 32;
const COLUMN_EMAIL_SIZE = 255;

const Row = struct {
    const ID_SIZE = @sizeOf(u32);
    const USERNAME_SIZE = @sizeOf([COLUMN_USERNAME_SIZE]u8);
    const EMAIL_SIZE = @sizeOf([COLUMN_EMAIL_SIZE]u8);

    id: u32,
    username: [COLUMN_USERNAME_SIZE]u8, // Fixed-size array for username
    email: [COLUMN_EMAIL_SIZE]u8, // Fixed-size array for email

    pub fn printRow(self: *Row) void {
        // Convert username and email to slices and stop at null-terminators
        const username = subArrayUntil(self.username, 0);
        const email = subArrayUntil(self.email, 0);

        // Print the row data
        std.debug.print("({}, {}, {})\n", .{ self.id, username, email });
    }
};

const ID_OFFSET: usize = 0;
const USERNAME_OFFSET: usize = ID_OFFSET + Row.ID_SIZE;
const EMAIL_OFFSET: usize = USERNAME_OFFSET + Row.USERNAME_SIZE;
const ROW_SIZE: usize = Row.ID_SIZE + Row.USERNAME_SIZE + Row.EMAIL_SIZE;

// Define page and table constants
const PAGE_SIZE: usize = 4096;
const TABLE_MAX_PAGES: usize = 100;
const ROWS_PER_PAGE: usize = PAGE_SIZE / ROW_SIZE;
const TABLE_MAX_ROWS: usize = ROWS_PER_PAGE * TABLE_MAX_PAGES;

pub fn serializeRow(source: *const Row, destination: []u8) void {
    std.mem.copy(u8, destination[ID_OFFSET .. ID_OFFSET + Row.ID_SIZE], source.id[0..Row.ID_SIZE]);
    std.mem.copy(u8, destination[USERNAME_OFFSET .. USERNAME_OFFSET + Row.USERNAME_SIZE], source.username);
    std.mem.copy(u8, destination[EMAIL_OFFSET .. EMAIL_OFFSET + Row.EMAIL_SIZE], source.email);
}

pub fn deserializeRow(source: []const u8, destination: *Row) void {
    destination.id = source[ID_OFFSET .. ID_OFFSET + Row.ID_SIZE];
    std.mem.copy(u8, destination.username, source[USERNAME_OFFSET .. USERNAME_OFFSET + Row.USERNAME_SIZE]);
    std.mem.copy(u8, destination.email, source[EMAIL_OFFSET .. EMAIL_OFFSET + Row.EMAIL_SIZE]);
}

pub const Table = struct {
    num_rows: u32,
    pages: [TABLE_MAX_PAGES]?*u8, // Array of pointers to pages (void* in C)

    pub fn new(allocator: *std.mem.Allocator) !*Table {
        var table = try allocator.create(Table); // Allocates memory for Table
        table.num_rows = 0;
        for (table.pages, 0..) |_, i| {
            table.pages[i] = null; // Set each page pointer to null
        }
        return table;
    }

    pub fn freeTable(self: *Table, allocator: *std.mem.Allocator) void {
        // Free pages and then the table itself
        for (self.pages) |page| {
            if (page) |p| {
                allocator.free(p);
            }
        }
        allocator.free(self);
    }

    pub fn rowSlot(self: *Table, row_num: u32) !*u8 {
        const page_num = row_num / ROWS_PER_PAGE;
        var page = self.pages[page_num];
        if (page == null) {
            // Allocate memory only when trying to access a page
            page = try self.allocPage(@intCast(page_num));
        }
        const row_offset = row_num % ROWS_PER_PAGE;
        const byte_offset = row_offset * ROW_SIZE;
        return page + byte_offset;
    }

    fn allocPage(self: *Table, page_num: u32) !*u8 {
        const allocator = std.heap.page_allocator;
        const page = try allocator.alloc(u8, PAGE_SIZE); // Allocate memory for the page
        self.pages[page_num] = @ptrCast(page.ptr); // Cast to non-nullable pointer and store
        return page.ptr;
    }

    pub fn executeInsert(self: *Table, statement: *Statement) ExecuteResult {
        if (self.num_rows >= TABLE_MAX_ROWS) {
            return ExecuteResult.TableFull;
        }

        // Get the row to insert and the row slot
        const row_to_insert = &statement.row_to_insert;
        const row_slot_ptr = self.rowSlot(self.num_rows);

        // Serialize the row into the appropriate slot in the table
        serializeRow(row_to_insert, row_slot_ptr);

        // Increment the number of rows in the table
        self.num_rows += 1;

        return ExecuteResult.success;
    }

    pub fn executeSelect(self: *Table) ExecuteResult {
        var row: Row = undefined;
        for (0..self.num_rows, 0..) |_, i| {
            // Get the row slot and deserialize it into the `row` struct
            const row_slot_ptr = self.rowSlot(@intCast(i));
            deserializeRow(row_slot_ptr, &row);

            // Print the row (printing code should be implemented)
            std.debug.print("Row {}: {} - {}\n", .{ row.id, row.username, row.email });
        }

        return ExecuteResult.success;
    }

    // Executes a prepared statement
    pub fn executeStatement(self: *Table, statement: *Statement) ExecuteResult {
        switch (statement.type) {
            StatementType.Insert => return self.executeInsert(statement),
            StatementType.Select => return self.executeSelect(),
        }
    }
};

pub fn subArrayUntil(data: []const u8, stop: u8) []const u8 {
    var i: usize = 0;
    for (data) |value| {
        if (value == stop) {
            return data[0..i];
        }
        i += 1;
    }
    return data;
}
