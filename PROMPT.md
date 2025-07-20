# TASK DESCRIPTION:
Perform a functional breakdown analysis on a single Go file, refactoring long functions into smaller, well-named private functions that improve code readability and maintainability. Exclude all test files from this refactoring analysis, test files are not eligible. Autonomously apply the refactoring to the existing codebase.

## CONTEXT:
You are acting as an automated Go code auditor specializing in functional decomposition. The goal is to identify functions exceeding reasonable complexity thresholds and refactor them into chains of smaller, purpose-specific functions. This improves code readability, testability, and maintainability while preserving all original functionality and error handling patterns.

Target metrics:
- Functions exceeding 50 lines of code (excluding comments and blank lines)
- Functions with cyclomatic complexity > 10
- Functions performing multiple distinct logical operations

## INSTRUCTIONS:
1. **File Selection Phase:**
   - Scan the provided Go codebase for files containing functions exceeding 50 lines
   - Prioritize files with the longest functions first
   - Select exactly one file for refactoring
   - If no functions exceed the threshold, skip to step 6

2. **Function Analysis Phase:**
   - Identify the longest function in the selected file
   - Map distinct logical tasks within this function by identifying:
     * Initialization/setup blocks
     * Validation/preprocessing steps
     * Core business logic segments
     * Error handling patterns
     * Cleanup/finalization operations
     * Loop bodies performing discrete operations
     * Conditional blocks with substantial logic

3. **Refactoring Design Phase:**
   - Plan extraction of each identified task into a private function
   - Design function signatures that:
     * Accept only necessary parameters
     * Return appropriate values including error types
     * Maintain the same error handling patterns as the original
   - Ensure extracted functions will be attached to the correct receiver (if methods)
   - Verify that variable scoping remains correct

4. **Implementation Phase:**
   - Extract each identified task into a private function following these naming conventions:
     * Use camelCase starting with lowercase letter
     * Begin with a verb describing the action
     * Be specific about the function's purpose
     * Examples: `validateUserInput()`, `calculateTaxRate()`, `buildResponseHeader()`
   - Add a 1-2 line comment above each new function following Go comment conventions:
     * Start with the function name
     * Describe what the function does, not how
     * Example: `// validateUserInput checks that all required fields are present and valid.`
   - Preserve all error handling:
     * Propagate errors up the call chain
     * Maintain original error wrapping/annotation
     * Keep defer statements in appropriate scope
   - Update the original function to call the new private functions in sequence

5. **Verification Phase:**
   - Confirm functional equivalence:
     * All original logic is preserved
     * Error handling paths remain identical
     * Return values are unchanged
   - Verify Go best practices:
     * Functions follow single responsibility principle
     * Error handling follows Go idioms
     * Variable scoping is correct
     * No unnecessary global state access

6. **Completion Phase:**
   - If refactoring was performed: Output message "Refactor complete: [filename] has been successfully decomposed."
   - If no refactoring needed: Output message "Refactor complete: No functions in the codebase require functional breakdown."

## FORMATTING REQUIREMENTS:
Present the refactored code using:
- Standard Go formatting (as produced by `go fmt`)
- Clear separation between the refactored main function and extracted helper functions
- Consistent indentation and spacing
- Proper placement of comments according to Go conventions

Structure your response as:
1. Brief analysis summary (2-3 sentences)
2. The complete refactored file with all changes
3. Completion message

## QUALITY CHECKS:
Before presenting the refactored code, verify:
- The refactored code compiles without errors
- All tests that passed before refactoring still pass
- No business logic has been altered
- Error handling is preserved exactly as in the original
- Each extracted function has a single, clear purpose
- Function names accurately describe their behavior
- Comments follow Go documentation standards
- No code duplication has been introduced
- Variable scoping is correct (no unintended closures or escapes)

## EXAMPLES:
Example of a function requiring breakdown:
```go
func processOrder(order Order) error {
    // 75 lines including validation, calculation, database operations, and notification
}
```

After refactoring:
```go
func processOrder(order Order) error {
    if err := validateOrder(order); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    
    total, err := calculateOrderTotal(order)
    if err != nil {
        return fmt.Errorf("calculation failed: %w", err)
    }
    
    if err := saveOrderToDatabase(order, total); err != nil {
        return fmt.Errorf("database save failed: %w", err)
    }
    
    if err := notifyCustomer(order); err != nil {
        // Non-critical error, log but don't fail
        log.Printf("customer notification failed: %v", err)
    }
    
    return nil
}

// validateOrder ensures all required order fields are present and valid.
// It returns an error if any validation rule is violated.
func validateOrder(order Order) error {
    // validation logic
}

// calculateOrderTotal computes the final price including tax and discounts.
// It returns the total amount or an error if calculation fails.
func calculateOrderTotal(order Order) (float64, error) {
    // calculation logic
}
```

#codebase

pkg/engine/game.go
pkg/config/config.go
pkg/event/event.go
pkg/physics/vector.go
pkg/physics/collision.go
pkg/physics/ship.go
pkg/network/client.go
pkg/network/server.go
pkg/entity/planet.go
pkg/entity/weapon.go
pkg/entity/entity.go
pkg/entity/ship.go
pkg/entity/renderer.go
pkg/render/terminal.go
pkg/render/renderer.go
