# AWSL (AWS Scripting Language) Specification

## Overview

AWSL is a domain-specific scripting language designed for interacting with AWS services, specifically Lambda and DynamoDB (for now). It provides a C-like syntax with built-in primitives for common AWS operations, eliminating the verbosity of shell scripts with AWS CLI and jq.

**Primary Goals:**
- Simplify AWS Lambda and DynamoDB operations
- Provide clean, readable syntax for cloud scripting
- Built-in output formatting (CSV, tables)
- First-class support for AWS contexts (profile, region)

**Implementation:**
- Interpreted language (not compiled)
- Written in Go
- Uses AWS SDK directly (not CLI wrapper)
- MIT licensed

---

## Lexical Structure

### Comments

```c
// Single-line comments only
// No multi-line comment blocks
```

### Identifiers

```
- Must start with a letter or underscore
- Can contain letters, digits, and underscores
- Case-sensitive
- Cannot be a reserved keyword

Valid:   foo, bar_baz, myVar123, _private, camelCase
Invalid: 123abc, my-var, for, true
```

### Keywords

```
fn       - Function declaration
true     - Boolean true literal
false    - Boolean false literal
null     - Null literal
if       - Conditional statement
else     - Conditional else branch
for      - Loop statement
in       - Iterator keyword (used with for)
return   - Return from function
profile  - AWS profile context setter
region   - AWS region context setter
```

### Operators

| Operator | Description |
|----------|-------------|
| `=` | Assignment |
| `+` | Addition |
| `-` | Subtraction |
| `*` | Multiplication |
| `/` | Division |
| `!` | Logical NOT |
| `==` | Equality |
| `!=` | Inequality |
| `<` | Less than |
| `>` | Greater than |
| `<=` | Less than or equal |
| `>=` | Greater than or equal |
| `.` | Member access |
| `|` | Pipe (for formatting) |

### Delimiters

| Symbol | Usage |
|--------|-------|
| `;` | Statement terminator |
| `:` | Key-value separator in objects and named arguments |
| `,` | Element separator |
| `(` `)` | Function calls, grouping, for loop syntax |
| `{` `}` | Blocks, object literals |
| `[` `]` | Array literals, index access |

---

## Types

### Primitive Types

| Type | Examples | Description |
|------|----------|-------------|
| `string` | `"hello"`, `"us-west-2"` | UTF-8 text, double-quoted |
| `int` | `42`, `0`, `-5` | 64-bit signed integer |
| `float` | `3.14`, `0.5` | 64-bit floating point |
| `bool` | `true`, `false` | Boolean value |
| `null` | `null` | Absence of value |

### Composite Types

| Type | Examples | Description |
|------|----------|-------------|
| `object` | `{name: "test", count: 5}` | Key-value map |
| `list` | `[1, 2, 3]`, `["a", "b"]` | Ordered collection |

### Type Coercion

- No implicit type coercion
- String concatenation with `+` requires both operands to be strings
- Comparison operators require matching types

---

## Syntax

### Variable Assignment

```c
name = "value";
count = 42;
enabled = true;
data = {key: "value", num: 123};
items = [1, 2, 3];
```

### Context Statements

```c
profile "profile-name";
region "aws-region";
```

These set the AWS context for subsequent operations. They are statements, not assignments.

### Objects

```c
// Object literal
user = {
    pk: "ORG#acme",
    sk: "USER#123",
    name: "Alice",
    active: true
};

// Member access
user.name;
user.active;

// Nested objects
config = {
    lambda: {
        memory: 256,
        timeout: 30
    }
};
config.lambda.memory;
```

### Lists

```c
// List literal
numbers = [1, 2, 3, 4, 5];
names = ["alice", "bob", "charlie"];

// Index access (zero-based)
numbers[0];  // 1
names[2];    // "charlie"

// Mixed types allowed
mixed = [1, "two", true, null];
```

### Control Flow

#### If Statement

```c
if (condition) {
    // statements
}

if (condition) {
    // statements
} else {
    // statements
}

if (x > 10) {
    print("large");
} else {
    print("small");
}
```

#### For Loop

```c
for (item in collection) {
    // statements using item
}

for (fn in functions) {
    print(fn.name);
}

for (i in [1, 2, 3]) {
    print(i);
}
```

### Functions

```c
// Declaration
fn functionName(param1, param2) {
    // statements
    return value;
}

// Call
result = functionName(arg1, arg2);

// Example
fn double(x) {
    return x * 2;
}

result = double(5);  // 10
```

### Named Arguments

For AWS service calls, named arguments are supported:

```c
// Named arguments use colon syntax
lambda.list(runtime: "python3.12");

dynamo.query(
    pk: "ORG#acme",
    sk_begins: "USER#",
    filter: {active: true}
);

// Positional and named can mix (positional first)
lambda.invoke("function-name", {payload: "data"});
```

### Pipe Operator

Used for output formatting:

```c
items | format csv;
items | format table;
```

---

## Built-in Functions

| Function | Description | Example |
|----------|-------------|---------|
| `print(...)` | Output values to stdout | `print("hello", x);` |
| `len(x)` | Length of string or list | `len([1,2,3])` → `3` |
| `type(x)` | Type of value as string | `type(42)` → `"int"` |

---

## AWS Service Bindings

### Lambda Namespace

```c
// List functions
functions = lambda.list();
functions = lambda.list(runtime: "python3.12");

// Get function details
fn = lambda.get("function-name");

// Invoke function
result = lambda.invoke("function-name", {
    user_id: "123",
    action: "process"
});

// Access invoke result
result.status_code;
result.payload;

// Function properties
fn.name;
fn.arn;
fn.runtime;
fn.memory;
fn.timeout;
fn.handler;
fn.last_modified;
```

### DynamoDB Namespace

```c
// Get table reference
users = dynamo.table("TableName");

// Query items
items = users.query(
    pk: "PARTITION_KEY",
    sk_begins: "PREFIX#",
    sk_between: ["START", "END"],
    filter: {active: true},
    limit: 100
);

// Get single item
item = users.get(pk: "PK_VALUE", sk: "SK_VALUE");

// Put item
users.put({
    pk: "ORG#acme",
    sk: "USER#456",
    name: "Bob",
    email: "bob@example.com"
});

// Delete item
users.delete(pk: "PK_VALUE", sk: "SK_VALUE");
users.delete(
    pk: "PK_VALUE",
    sk: "SK_VALUE",
    condition: {active: false}
);

// Scan (use sparingly)
all_items = users.scan();
all_items = users.scan(filter: {type: "admin"});
```

---

## Output Formatting

```c
// CSV output
items | format csv;

// Table output (ASCII formatted)
items | format table;

// Format specific fields
items | format table;
// Outputs:
// +--------+------------------+--------+
// | name   | email            | active |
// +--------+------------------+--------+
// | Alice  | alice@example.com| true   |
// | Bob    | bob@example.com  | false  |
// +--------+------------------+--------+
```

---

## Grammar (EBNF)

```ebnf
program        = { statement } ;

statement      = assignment
               | expr_statement
               | context_statement
               | if_statement
               | for_statement
               | return_statement
               | function_decl ;

context_statement = ( "profile" | "region" ) string ";" ;

assignment     = identifier "=" expr ";" ;

expr_statement = expr ";" ;

if_statement   = "if" "(" expr ")" block [ "else" block ] ;

for_statement  = "for" "(" identifier "in" expr ")" block ;

return_statement = "return" [ expr ] ";" ;

function_decl  = "fn" identifier "(" [ param_list ] ")" block ;

param_list     = identifier { "," identifier } ;

block          = "{" { statement } "}" ;

expr           = logic_or ;

logic_or       = logic_and { "||" logic_and } ;

logic_and      = equality { "&&" equality } ;

equality       = comparison { ( "==" | "!=" ) comparison } ;

comparison     = term { ( "<" | ">" | "<=" | ">=" ) term } ;

term           = factor { ( "+" | "-" ) factor } ;

factor         = unary { ( "*" | "/" ) unary } ;

unary          = ( "!" | "-" ) unary | postfix ;

postfix        = primary { call | index | member | pipe } ;

call           = "(" [ arg_list ] ")" ;

index          = "[" expr "]" ;

member         = "." identifier ;

pipe           = "|" "format" ( "csv" | "table" ) ;

arg_list       = arg { "," arg } ;

arg            = [ identifier ":" ] expr ;

primary        = identifier
               | number
               | string
               | "true" | "false" | "null"
               | "(" expr ")"
               | list_literal
               | object_literal ;

list_literal   = "[" [ expr { "," expr } ] "]" ;

object_literal = "{" [ pair { "," pair } ] "}" ;

pair           = identifier ":" expr ;

identifier     = letter { letter | digit | "_" } ;

number         = digit { digit } [ "." digit { digit } ] ;

string         = '"' { character } '"' ;

letter         = "a"..."z" | "A"..."Z" | "_" ;

digit          = "0"..."9" ;
```

---

## Token Types

```
// Special
ILLEGAL, EOF

// Identifiers and literals
IDENT, INT, FLOAT, STRING

// Operators
ASSIGN (=), PLUS (+), MINUS (-), BANG (!), ASTERISK (*), SLASH (/)
LT (<), GT (>), EQ (==), NOT_EQ (!=), LTE (<=), GTE (>=)

// Delimiters
COMMA (,), SEMICOLON (;), COLON (:), DOT (.), PIPE (|)
LPAREN ((), RPAREN ()), LBRACE ({), RBRACE (})
LBRACKET ([), RBRACKET (])

// Keywords
FUNCTION (fn), TRUE (true), FALSE (false), NULL (null)
IF (if), ELSE (else), FOR (for), IN (in), RETURN (return)
PROFILE (profile), REGION (region)
```

---

## Complete Example Script

```c
// Configure AWS context
profile "production";
region "us-west-2";

// List all Python Lambda functions
functions = lambda.list(runtime: "python3.12");

// Print summary
print("Found", len(functions), "functions");

// Iterate and display
for (fn in functions) {
    print(fn.name, fn.memory, fn.timeout);
}

// Query DynamoDB for active users
users_table = dynamo.table("Users");

active_users = users_table.query(
    pk: "ORG#acme",
    sk_begins: "USER#",
    filter: {active: true}
);

// Output as table
active_users | format table;

// Invoke a Lambda function
result = lambda.invoke("process-user", {
    user_id: "12345",
    action: "notify"
});

if (result.status_code == 200) {
    print("Success:", result.payload);
} else {
    print("Error:", result.status_code);
}

// Define reusable function
fn get_active_users(org_id) {
    table = dynamo.table("Users");
    return table.query(
        pk: org_id,
        sk_begins: "USER#",
        filter: {active: true}
    );
}

admins = get_active_users("ORG#admin");
admins | format csv;
```

---

## Error Handling

Runtime errors include:
- AWS API errors (permissions, resource not found)
- Type errors (invalid operations)
- Reference errors (undefined variables)

Errors display:
```
Error at line 15, column 8: undefined variable 'foo'
Error at line 22: AWS error: ResourceNotFoundException: Table 'Users' not found
```

---

## Reserved for Future

These features are not in v1 but may be added:

- `wait` keyword for polling resource states
- `dry_run` blocks for safe testing
- Time literals (`30 days`, `5 min`)
- S3 namespace
- EC2 namespace
- Try/catch error handling
- String interpolation

---

## Usage

```bash
# Run a script
awsl script.awsl

# Show version
awsl --version
```
