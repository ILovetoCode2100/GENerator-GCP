# Virtuoso CLI Command Structure Mapping

## Unified Command Pattern

All commands follow this standardized structure:

```
api-cli <category> <subcommand> [checkpoint-id] <arg1> <arg2> ... <arg5> [position] [--flag value]
```

### Components:

- **Category**: High-level action group (assert, interact, navigate, etc.)
- **Subcommand**: Specific operation (exists, click, to, etc.)
- **[checkpoint-id]**: Optional positional argument (falls back to session context)
- **<arg1-5>**: Required positional arguments (up to 5)
- **[position]**: Optional step position (always last positional arg)
- **[--flag value]**: Optional modifiers (reserved for non-core parameters)

## Command Structure Mapping

### 1. Assert Commands (12 types)

**Pattern**: `assert <type> [checkpoint-id] <args...> [position]`

| Subcommand | Arg1     | Arg2    | Arg3 | Arg4 | Arg5 | Flags |
| ---------- | -------- | ------- | ---- | ---- | ---- | ----- |
| exists     | element  | -       | -    | -    | -    | -     |
| not-exists | element  | -       | -    | -    | -    | -     |
| equals     | element  | value   | -    | -    | -    | -     |
| not-equals | element  | value   | -    | -    | -    | -     |
| checked    | element  | -       | -    | -    | -    | -     |
| selected   | element  | -       | -    | -    | -    | -     |
| gt         | element  | value   | -    | -    | -    | -     |
| gte        | element  | value   | -    | -    | -    | -     |
| lt         | element  | value   | -    | -    | -    | -     |
| lte        | element  | value   | -    | -    | -    | -     |
| matches    | element  | pattern | -    | -    | -    | -     |
| variable   | variable | value   | -    | -    | -    | -     |

### 2. Interact Commands (6 types)

**Pattern**: `interact <action> [checkpoint-id] <selector> [value] [position]`

| Subcommand   | Arg1     | Arg2 | Arg3 | Arg4 | Arg5 | Flags             |
| ------------ | -------- | ---- | ---- | ---- | ---- | ----------------- |
| click        | selector | -    | -    | -    | -    | --position (enum) |
| double-click | selector | -    | -    | -    | -    | -                 |
| right-click  | selector | -    | -    | -    | -    | -                 |
| hover        | selector | -    | -    | -    | -    | -                 |
| write        | selector | text | -    | -    | -    | -                 |
| key          | selector | key  | -    | -    | -    | --modifiers       |

### 3. Navigate Commands (10 types)

**Pattern**: `navigate <action> [checkpoint-id] <args...> [position]`

| Subcommand      | Arg1     | Arg2 | Arg3 | Arg4 | Arg5 | Flags    |
| --------------- | -------- | ---- | ---- | ---- | ---- | -------- |
| to              | url      | -    | -    | -    | -    | -        |
| scroll-top      | -        | -    | -    | -    | -    | -        |
| scroll-bottom   | -        | -    | -    | -    | -    | -        |
| scroll-element  | selector | -    | -    | -    | -    | -        |
| scroll-position | x,y      | -    | -    | -    | -    | -        |
| scroll-by       | x,y      | -    | -    | -    | -    | --x, --y |
| scroll-up       | -        | -    | -    | -    | -    | -        |
| scroll-down     | -        | -    | -    | -    | -    | -        |

### 4. Data Commands (6 types)

**Pattern**: `data <operation> <subtype> [checkpoint-id] <args...> [position]`

| Subcommand         | Arg1     | Arg2      | Arg3     | Arg4 | Arg5 | Flags                                   |
| ------------------ | -------- | --------- | -------- | ---- | ---- | --------------------------------------- |
| store element-text | selector | variable  | -        | -    | -    | -                                       |
| store literal      | value    | variable  | -        | -    | -    | -                                       |
| store attribute    | selector | attribute | variable | -    | -    | -                                       |
| cookie create      | name     | value     | -        | -    | -    | --domain, --path, --secure, --http-only |
| cookie delete      | name     | -         | -        | -    | -    | -                                       |
| cookie clear-all   | -        | -         | -        | -    | -    | -                                       |

### 5. Dialog Commands (4 types)

**Pattern**: `dialog <type> [checkpoint-id] <action> [position]`

| Subcommand      | Arg1 | Arg2 | Arg3 | Arg4 | Arg5 | Flags |
| --------------- | ---- | ---- | ---- | ---- | ---- | ----- |
| alert dismiss   | -    | -    | -    | -    | -    | -     |
| confirm dismiss | -    | -    | -    | -    | -    | -     |
| prompt dismiss  | -    | -    | -    | -    | -    | -     |
| prompt dismiss  | text | -    | -    | -    | -    | -     |

### 6. Wait Commands (3 types)

**Pattern**: `wait <type> [checkpoint-id] <args...> [position]`

| Subcommand          | Arg1         | Arg2 | Arg3 | Arg4 | Arg5 | Flags |
| ------------------- | ------------ | ---- | ---- | ---- | ---- | ----- |
| element             | selector     | -    | -    | -    | -    | -     |
| element-not-visible | selector     | -    | -    | -    | -    | -     |
| time                | milliseconds | -    | -    | -    | -    | -     |

### 7. Window Commands (5 types)

**Pattern**: `window <operation> [checkpoint-id] <args...> [position]`

| Subcommand          | Arg1         | Arg2  | Arg3 | Arg4 | Arg5 | Flags |
| ------------------- | ------------ | ----- | ---- | ---- | ---- | ----- |
| resize              | WIDTHxHEIGHT | -     | -    | -    | -    | -     |
| maximize            | -            | -     | -    | -    | -    | -     |
| switch tab          | type         | value | -    | -    | -    | -     |
| switch iframe       | selector     | -     | -    | -    | -    | -     |
| switch parent-frame | -            | -     | -    | -    | -    | -     |

### 8. Mouse Commands (6 types)

**Pattern**: `mouse <action> [checkpoint-id] <args...> [position]`

| Subcommand | Arg1     | Arg2 | Arg3 | Arg4 | Arg5 | Flags |
| ---------- | -------- | ---- | ---- | ---- | ---- | ----- |
| move-to    | selector | -    | -    | -    | -    | -     |
| move-by    | x,y      | -    | -    | -    | -    | -     |
| move       | x,y      | -    | -    | -    | -    | -     |
| down       | -        | -    | -    | -    | -    | -     |
| up         | -        | -    | -    | -    | -    | -     |
| enter      | selector | -    | -    | -    | -    | -     |

### 9. Select Commands (3 types)

**Pattern**: `select <type> [checkpoint-id] <selector> <value> [position]`

| Subcommand | Arg1     | Arg2  | Arg3 | Arg4 | Arg5 | Flags |
| ---------- | -------- | ----- | ---- | ---- | ---- | ----- |
| option     | selector | value | -    | -    | -    | -     |
| index      | selector | index | -    | -    | -    | -     |
| last       | selector | -     | -    | -    | -    | -     |

### 10. File Commands (2 types)

**Pattern**: `file <action> [checkpoint-id] <selector> <url> [position]`

| Subcommand | Arg1     | Arg2 | Arg3 | Arg4 | Arg5 | Flags |
| ---------- | -------- | ---- | ---- | ---- | ---- | ----- |
| upload     | selector | url  | -    | -    | -    | -     |
| upload-url | selector | url  | -    | -    | -    | -     |

### 11. Misc Commands (2 types)

**Pattern**: `misc <type> [checkpoint-id] <args...> [position]`

| Subcommand | Arg1       | Arg2 | Arg3 | Arg4 | Arg5 | Flags |
| ---------- | ---------- | ---- | ---- | ---- | ---- | ----- |
| comment    | text       | -    | -    | -    | -    | -     |
| execute    | javascript | -    | -    | -    | -    | -     |

### 12. Library Commands (6 types)

**Pattern**: `library <action> <args...>`

| Subcommand  | Arg1          | Arg2       | Arg3     | Arg4 | Arg5 | Flags |
| ----------- | ------------- | ---------- | -------- | ---- | ---- | ----- |
| add         | checkpoint-id | -          | -        | -    | -    | -     |
| get         | library-id    | -          | -        | -    | -    | -     |
| attach      | journey-id    | library-id | position | -    | -    | -     |
| move-step   | library-id    | from-pos   | to-pos   | -    | -    | -     |
| remove-step | library-id    | position   | -        | -    | -    | -     |
| update      | library-id    | title      | -        | -    | -    | -     |

## Implementation Status

### ✅ Already Using BaseCommand Pattern (4 groups)

- interact
- navigate
- file
- misc

### 🔄 Migrated to BaseCommand Pattern (2 groups)

- assert (assert_v2.go created)
- data (data_v2.go created)

### ⏳ Pending Migration (5 groups)

- window
- wait
- mouse
- dialog
- select

### 🎯 Special Cases

- library (doesn't need checkpoint/position pattern)

## Key Benefits of Standardization

1. **Consistency**: All commands follow the same positional argument pattern
2. **Session Support**: Automatic checkpoint resolution from session context
3. **Position Handling**: Consistent last-position argument with auto-increment
4. **AI-Friendly**: Predictable structure for programmatic generation
5. **User-Friendly**: Easier to learn and remember patterns

## Migration Notes

When migrating commands to the new pattern:

1. Replace `--checkpoint` flag with positional argument
2. Use `BaseCommand.ResolveCheckpointAndPosition()` method
3. Ensure position is always the last positional argument
4. Keep flags only for optional modifiers (not core arguments)
5. Maintain backward compatibility during transition
