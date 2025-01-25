const std = @import("std");

pub const MetaCommandResult = enum {
    Success,
    UnrecognizedCommand,
};

pub const PrepareResult = enum {
    Success,
    UnrecognizedStatement,
};

pub const StatementType = enum {
    Insert,
    Select,
};

pub const Statement = struct {
    typ: StatementType,
};
