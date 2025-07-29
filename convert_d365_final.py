#!/usr/bin/env python3
"""
Final D365 to Virtuoso YAML converter - simplified format
"""
import re
import os
import yaml
from pathlib import Path


class D365SimplifiedConverter:
    """Convert D365 NLP test format to Virtuoso simplified YAML format"""

    def __init__(self):
        self.command_mappings = {
            # Navigation
            r'Navigate to "([^"]+)"': lambda m: {"navigate": m.group(1)},
            r"Navigate to (.+)": lambda m: {"navigate": m.group(1).strip()},
            # Click actions
            r'Click on "([^"]+)"': lambda m: {"click": m.group(1)},
            r"Click on (.+)": lambda m: {"click": m.group(1).strip()},
            r'Double click on "([^"]+)"': lambda m: {"double_click": m.group(1)},
            # Write/Input actions
            r'Write "([^"]+)" in field "([^"]+)"': lambda m: {
                "write": {"selector": m.group(2), "text": m.group(1)}
            },
            r'Write (.+) in field "([^"]+)"': lambda m: {
                "write": {"selector": m.group(2), "text": m.group(1).strip()}
            },
            r'Write "([^"]+)" in field (.+)': lambda m: {
                "write": {"selector": m.group(2).strip(), "text": m.group(1)}
            },
            # Select/Pick actions
            r'Pick "([^"]+)" from "([^"]+)"': lambda m: {
                "select": {"selector": m.group(2), "option": m.group(1)}
            },
            r'Pick "([^"]+)" from dropdown "([^"]+)"': lambda m: {
                "select": {"selector": m.group(2), "option": m.group(1)}
            },
            # Wait actions
            r"Wait (\d+) seconds?": lambda m: {"wait": int(m.group(1))},
            # Assertions
            r'Assert that element "([^"]+)" exists': lambda m: {"assert": m.group(1)},
            r'Assert that element "([^"]+)" exists on page': lambda m: {
                "assert": m.group(1)
            },
            r'Assert that element "([^"]+)" equals "([^"]+)"': lambda m: {
                "assert_text": {"selector": m.group(1), "expected": m.group(2)}
            },
            r'Assert that element "([^"]+)" equals (.+)': lambda m: {
                "assert_text": {"selector": m.group(1), "expected": m.group(2).strip()}
            },
            r'Assert that element "([^"]+)" shows "([^"]+)"': lambda m: {
                "assert_text": {"selector": m.group(1), "expected": m.group(2)}
            },
            # Store actions
            r'Store element text of "([^"]+)" in (.+)': lambda m: {
                "store": {
                    "type": "text",
                    "selector": m.group(1),
                    "variable": m.group(2).strip(),
                }
            },
            r'Store value "([^"]+)" in (.+)': lambda m: {
                "store": {
                    "type": "value",
                    "value": m.group(1),
                    "variable": m.group(2).strip(),
                }
            },
            r'Store element value of "([^"]+)" in (.+)': lambda m: {
                "store": {
                    "type": "element_value",
                    "selector": m.group(1),
                    "variable": m.group(2).strip(),
                }
            },
            # Other actions
            r'Upload "([^"]+)" to "([^"]+)"': lambda m: {
                "upload": {"file": m.group(1), "selector": m.group(2)}
            },
            r"Press (.+) key": lambda m: {"key": m.group(1).strip()},
            r"Press (.+)": lambda m: {"key": m.group(1).strip()},
            r'Scroll to "([^"]+)"': lambda m: {"scroll": m.group(1)},
            r"Switch to next tab": lambda m: {"switch_tab": "next"},
            r"Switch to previous tab": lambda m: {"switch_tab": "previous"},
            r'Look for element "([^"]+)" on page': lambda m: {"wait_for": m.group(1)},
            r"Execute script: (.+)": lambda m: {"execute": m.group(1).strip()},
            r'Mouse drag "([^"]+)" to "([^"]+)"': lambda m: {
                "drag": {"from": m.group(1), "to": m.group(2)}
            },
        }

    def parse_test_file(self, file_path):
        """Parse a D365 test file and extract test suites and cases"""
        with open(file_path, "r", encoding="utf-8") as f:
            content = f.read()

        # Extract module name from file name
        module_name = (
            Path(file_path)
            .stem.replace("D365_", "")
            .replace("_Tests", "")
            .replace("_", " ")
        )

        # Parse test suites
        test_suites = []
        suite_pattern = r"## Test Suite: (.+?)\n(.*?)(?=## Test Suite:|## Test Data Variables|## Performance Validations|## Cleanup Steps|$)"

        for suite_match in re.finditer(suite_pattern, content, re.DOTALL):
            suite_name = suite_match.group(1).strip()
            suite_content = suite_match.group(2)

            # Parse test cases within suite
            test_cases = []
            case_pattern = r"### Test Case: (.+?)\n// Test Description: (.+?)\n(.*?)(?=### Test Case:|## |$)"

            for case_match in re.finditer(case_pattern, suite_content, re.DOTALL):
                test_id = case_match.group(1).strip()
                description = case_match.group(2).strip()
                steps_content = case_match.group(3).strip()

                # Parse steps
                steps = self.parse_steps(steps_content)

                test_cases.append(
                    {"id": test_id, "description": description, "steps": steps}
                )

            test_suites.append({"name": suite_name, "test_cases": test_cases})

        return {"module_name": module_name, "test_suites": test_suites}

    def parse_steps(self, steps_content):
        """Parse test steps from content into simplified format"""
        steps = []
        lines = steps_content.strip().split("\n")

        for line in lines:
            line = line.strip()
            if not line or line.startswith("//"):
                continue

            # Try to match against command patterns
            step_parsed = False
            for pattern, mapper in self.command_mappings.items():
                match = re.match(pattern, line)
                if match:
                    step_data = mapper(match)
                    steps.append(step_data)
                    step_parsed = True
                    break

            if not step_parsed and line:
                # If no pattern matches, add as a comment
                steps.append({"comment": line})

        return steps

    def convert_to_yaml(self, test_data, output_dir):
        """Convert parsed test data to simplified YAML files"""
        module_name = test_data["module_name"]

        # Map module names to directory names
        dir_mapping = {
            "Sales Module": "sales",
            "Customer Service": "customer-service",
            "Field Service": "field-service",
            "Marketing": "marketing",
            "Finance Operations": "finance-operations",
            "Project Operations": "project-operations",
            "Human Resources": "human-resources",
            "Supply Chain": "supply-chain",
            "Commerce": "commerce",
        }

        module_output_dir = os.path.join(
            output_dir,
            dir_mapping.get(module_name, module_name.lower().replace(" ", "-")),
        )
        os.makedirs(module_output_dir, exist_ok=True)

        yaml_files = []

        # Create one master YAML file per module with all test cases
        module_yaml_file = os.path.join(
            module_output_dir,
            f"{dir_mapping.get(module_name, module_name.lower().replace(' ', '-'))}-all-tests.yaml",
        )

        # Also create individual test case files
        for suite in test_data["test_suites"]:
            for test_case in suite["test_cases"]:
                test_case_name_clean = (
                    re.sub(r"[^\w\s-]", "", test_case["id"])
                    .strip()
                    .replace(" ", "-")
                    .lower()
                )
                yaml_file_path = os.path.join(
                    module_output_dir, f"{test_case_name_clean}.yaml"
                )

                # Create simplified YAML structure
                yaml_data = {
                    "name": f"{test_case['id']} - {test_case['description']}",
                    "description": f"{test_case['description']} (Module: {module_name}, Suite: {suite['name']})",
                    "starting_url": "https://[instance].crm.dynamics.com",
                    "steps": test_case["steps"],
                }

                # Write YAML file with proper formatting
                with open(yaml_file_path, "w", encoding="utf-8") as f:
                    yaml.dump(
                        yaml_data,
                        f,
                        default_flow_style=False,
                        sort_keys=False,
                        allow_unicode=True,
                        width=1000,
                        indent=2,
                    )

                yaml_files.append(yaml_file_path)
                print(f"Created: {yaml_file_path}")

        return yaml_files

    def convert_all_files(self, input_dir, output_dir):
        """Convert all D365 test files in a directory"""
        test_files = [
            f
            for f in os.listdir(input_dir)
            if f.endswith(".txt") and f.startswith("D365_")
        ]

        all_yaml_files = []

        for test_file in test_files:
            print(f"\nProcessing: {test_file}")
            file_path = os.path.join(input_dir, test_file)

            try:
                # Parse test file
                test_data = self.parse_test_file(file_path)

                # Convert to YAML
                yaml_files = self.convert_to_yaml(test_data, output_dir)
                all_yaml_files.extend(yaml_files)

            except Exception as e:
                print(f"Error processing {test_file}: {str(e)}")
                import traceback

                traceback.print_exc()

        return all_yaml_files


def main():
    """Main conversion function"""
    converter = D365SimplifiedConverter()

    input_dir = "/Users/marklovelady/Downloads/D365"
    output_dir = "d365-virtuoso-tests-final"

    print("Starting D365 test conversion (simplified format)...")
    yaml_files = converter.convert_all_files(input_dir, output_dir)

    print(f"\nConversion complete! Created {len(yaml_files)} YAML files.")
    print("\nNext steps:")
    print(
        "1. Review the generated YAML files in the d365-virtuoso-tests-final directory"
    )
    print("2. Update instance URLs in the YAML files")
    print("3. Deploy tests using: ./bin/api-cli run-test <yaml-file>")
    print("4. Or deploy all tests using the deployment script")


if __name__ == "__main__":
    main()
