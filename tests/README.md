# HTMX Instance Rendering Tests

This directory contains comprehensive tests for the HTMX instance rendering functionality.

## Test Files

### 1. `simple_htmx_test.go` - Core Functionality Tests
Tests the fundamental HTMX functionality without server complexity:

- **InstanceNamesGenerated**: Verifies instance names are generated correctly
- **HTMLIDFormat**: Ensures instance IDs follow the correct format (`instance-{AutoName}`)
- **HTMXOOBResponseFormat**: Tests out-of-band update response format
- **ProgressPollingHTMLFormat**: Verifies progress polling HTML structure
- **CompletionHTMLFormat**: Tests completion HTML without HTMX attributes

### 2. `TestHTMXAttributes` - HTMX Attribute Tests
Tests HTMX attribute combinations:

- **ProgressPollingAttributes**: Verifies progress polling has correct HTMX attributes
- **OOBUpdateAttributes**: Tests OOB updates have correct HTMX attributes
- **InstanceCardAttributes**: Ensures instance cards have correct IDs without conflicting HTMX attributes

## Running the Tests

### Run All Tests
```bash
go test ./tests -v
```

### Run Specific Test
```bash
# Run core functionality tests
go test ./tests -run TestInstanceRenderingSimple -v

# Run HTMX attributes tests
go test ./tests -run TestHTMXAttributes -v

# Run all HTMX tests
go test ./tests -run "TestInstanceRenderingSimple|TestHTMXAttributes" -v
```

## What the Tests Verify

### ✅ **Core Functionality**
- Instance names are generated correctly
- HTML IDs follow the correct format (`instance-{AutoName}`)
- HTMX OOB response format is correct
- Progress polling HTML has required HTMX attributes
- Completion HTML doesn't have HTMX attributes

### ✅ **HTMX Attributes**
- Progress polling has correct `hx-get`, `hx-trigger`, `hx-swap` attributes
- OOB updates have correct `hx-swap-oob` and target attributes
- Instance cards have correct IDs without conflicting HTMX attributes

## Test Scenarios Covered

### 1. **Config Page Rendering**
- Navigation to config file page renders a div for each InstanceConfig
- Each div has an ID with the format `instance-{AutoName}`
- All expected instances are present

### 2. **HTMX OOB Updates**
- After clicking process, processing starts correctly
- Instance cards are added to the instances grid via HTMX OOB updates
- All instance configs are replaced with processing results
- Each instance maintains its correct ID throughout the process
- Polling stops when processing is complete

## Expected Behavior

1. **Config Page**: Shows instance divs with IDs like `instance-small-tray-box`, `instance-large-tray-35-box`, etc.

2. **Processing**: Shows progress updates and gradually adds instance cards to the grid

3. **Completion**: Has all instance cards in the grid with correct IDs and stops polling

## Debugging

If tests fail, check:
- Config file exists at `../examples/small-tray/config.toml` (relative to tests directory)
- All required HTMX attributes are present in responses
- Instance IDs follow the correct format
- OOB responses target the correct elements

## Test Results

The tests verify that:
- ✅ Instance rendering works correctly
- ✅ HTMX OOB updates function properly
- ✅ Progress polling uses correct HTMX attributes
- ✅ Completion stops polling appropriately
- ✅ Instance cards are placed in the correct grid, not the header nav
