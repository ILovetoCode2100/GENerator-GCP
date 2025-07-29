#!/usr/bin/env python3
"""
Convert D365 Virtuoso NLP test files to YAML format
"""
import re
import os
import yaml
from pathlib import Path


class D365TestConverter:
    """Convert D365 NLP test format to Virtuoso YAML format"""

    def __init__(self):
        self.command_mappings = {
            r'Navigate to "([^"]+)"': lambda m: {
                "stepType": "navigate",
                "url": m.group(1),
            },
            r"Navigate to (.+)": lambda m: {
                "stepType": "navigate",
                "url": m.group(1).strip(),
            },
            r'Click on "([^"]+)"': lambda m: {
                "stepType": "click",
                "target": m.group(1),
            },
            r"Click on (.+)": lambda m: {
                "stepType": "click",
                "target": m.group(1).strip(),
            },
            r'Write "([^"]+)" in field "([^"]+)"': lambda m: {
                "stepType": "write",
                "target": m.group(2),
                "text": m.group(1),
            },
            r'Write (.+) in field "([^"]+)"': lambda m: {
                "stepType": "write",
                "target": m.group(2),
                "text": m.group(1).strip(),
            },
            r'Write "([^"]+)" in field (.+)': lambda m: {
                "stepType": "write",
                "target": m.group(2).strip(),
                "text": m.group(1),
            },
            r'Pick "([^"]+)" from "([^"]+)"': lambda m: {
                "stepType": "selectOption",
                "target": m.group(2),
                "option": m.group(1),
            },
            r'Pick "([^"]+)" from dropdown "([^"]+)"': lambda m: {
                "stepType": "selectOption",
                "target": m.group(2),
                "option": m.group(1),
            },
            r"Wait (\d+) seconds?": lambda m: {
                "stepType": "wait",
                "seconds": int(m.group(1)),
            },
            r'Assert that element "([^"]+)" exists': lambda m: {
                "stepType": "assertElement",
                "target": m.group(1),
                "exists": True,
            },
            r'Assert that element "([^"]+)" exists on page': lambda m: {
                "stepType": "assertElement",
                "target": m.group(1),
                "exists": True,
            },
            r'Assert that element "([^"]+)" equals "([^"]+)"': lambda m: {
                "stepType": "assertText",
                "target": m.group(1),
                "expectedText": m.group(2),
            },
            r'Assert that element "([^"]+)" equals (.+)': lambda m: {
                "stepType": "assertText",
                "target": m.group(1),
                "expectedText": m.group(2).strip(),
            },
            r'Assert that element "([^"]+)" shows "([^"]+)"': lambda m: {
                "stepType": "assertText",
                "target": m.group(1),
                "expectedText": m.group(2),
            },
            r'Assert that element "([^"]+)" is greater than (.+)': lambda m: {
                "stepType": "assertValue",
                "target": m.group(1),
                "operator": "greaterThan",
                "expectedValue": m.group(2).strip(),
            },
            r'Store element text of "([^"]+)" in (.+)': lambda m: {
                "stepType": "storeText",
                "target": m.group(1),
                "variable": m.group(2).strip(),
            },
            r'Store value "([^"]+)" in (.+)': lambda m: {
                "stepType": "storeValue",
                "value": m.group(1),
                "variable": m.group(2).strip(),
            },
            r"Store value (.+) in (.+)": lambda m: {
                "stepType": "storeValue",
                "value": m.group(1).strip(),
                "variable": m.group(2).strip(),
            },
            r'Store element value of "([^"]+)" in (.+)': lambda m: {
                "stepType": "storeElementValue",
                "target": m.group(1),
                "variable": m.group(2).strip(),
            },
            r'Upload "([^"]+)" to "([^"]+)"': lambda m: {
                "stepType": "uploadFile",
                "filePath": m.group(1),
                "target": m.group(2),
            },
            r"Press (.+) key": lambda m: {
                "stepType": "pressKey",
                "key": m.group(1).strip(),
            },
            r"Press (.+)": lambda m: {
                "stepType": "pressKey",
                "key": m.group(1).strip(),
            },
            r'Scroll to "([^"]+)"': lambda m: {
                "stepType": "scrollTo",
                "target": m.group(1),
            },
            r"Scroll to (.+)": lambda m: {
                "stepType": "scrollTo",
                "target": m.group(1).strip(),
            },
            r'Double click on "([^"]+)"': lambda m: {
                "stepType": "doubleClick",
                "target": m.group(1),
            },
            r'Mouse drag "([^"]+)" to "([^"]+)"': lambda m: {
                "stepType": "dragAndDrop",
                "source": m.group(1),
                "target": m.group(2),
            },
            r"Switch to next tab": lambda m: {
                "stepType": "switchTab",
                "direction": "next",
            },
            r"Switch to previous tab": lambda m: {
                "stepType": "switchTab",
                "direction": "previous",
            },
            r'Look for element "([^"]+)" on page': lambda m: {
                "stepType": "lookForElement",
                "target": m.group(1),
            },
            r"Execute script: (.+)": lambda m: {
                "stepType": "executeScript",
                "script": m.group(1).strip(),
            },
            r'If element "([^"]+)" exists then (.+)': lambda m: {
                "stepType": "conditional",
                "condition": {"element": m.group(1), "exists": True},
                "action": m.group(2).strip(),
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

        # Extract test data variables
        variables = {}
        var_pattern = r"## Test Data Variables\n(.*?)(?=## |$)"
        var_match = re.search(var_pattern, content, re.DOTALL)
        if var_match:
            var_content = var_match.group(1)
            for line in var_content.strip().split("\n"):
                if "=" in line:
                    key, value = line.split("=", 1)
                    variables[key.strip()] = value.strip()

        return {
            "module_name": module_name,
            "test_suites": test_suites,
            "variables": variables,
        }

    def parse_steps(self, steps_content):
        """Parse test steps from content"""
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
                steps.append({"stepType": "comment", "text": line})

        return steps

    def convert_to_yaml(self, test_data, output_dir):
        """Convert parsed test data to YAML files"""
        module_name = test_data["module_name"]
        module_dir = module_name.lower().replace(" ", "-")

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
            output_dir, dir_mapping.get(module_name, module_dir)
        )
        os.makedirs(module_output_dir, exist_ok=True)

        yaml_files = []

        # Create a YAML file for each test suite
        for suite in test_data["test_suites"]:
            suite_name_clean = (
                re.sub(r"[^\w\s-]", "", suite["name"]).strip().replace(" ", "-").lower()
            )

            # Create a YAML file for each test case in extended format
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

                # Convert steps to extended format
                extended_steps = []
                for step in test_case["steps"]:
                    if step["stepType"] == "navigate":
                        extended_steps.append(
                            {"type": "navigate", "target": step["url"]}
                        )
                    elif step["stepType"] == "click":
                        extended_steps.append(
                            {
                                "type": "interact",
                                "command": "click",
                                "target": step["target"],
                            }
                        )
                    elif step["stepType"] == "write":
                        extended_steps.append(
                            {
                                "type": "interact",
                                "command": "write",
                                "target": step["target"],
                                "value": step["text"],
                            }
                        )
                    elif step["stepType"] == "wait":
                        extended_steps.append(
                            {
                                "type": "wait",
                                "command": "time",
                                "value": step["seconds"],
                            }
                        )
                    elif step["stepType"] == "assertElement":
                        extended_steps.append(
                            {
                                "type": "assert",
                                "command": "exists"
                                if step.get("exists", True)
                                else "not_exists",
                                "target": step["target"],
                            }
                        )
                    elif step["stepType"] == "assertText":
                        extended_steps.append(
                            {
                                "type": "assert",
                                "command": "text",
                                "target": step["target"],
                                "expected": step["expectedText"],
                            }
                        )
                    elif step["stepType"] == "storeText":
                        extended_steps.append(
                            {
                                "type": "store",
                                "command": "text",
                                "target": step["target"],
                                "variable": step["variable"].replace("$", ""),
                            }
                        )
                    elif step["stepType"] == "storeValue":
                        extended_steps.append(
                            {
                                "type": "store",
                                "command": "value",
                                "value": step["value"],
                                "variable": step["variable"].replace("$", ""),
                            }
                        )
                    elif step["stepType"] == "selectOption":
                        extended_steps.append(
                            {
                                "type": "interact",
                                "command": "select",
                                "target": step["target"],
                                "value": step["option"],
                            }
                        )
                    elif step["stepType"] == "comment":
                        # Skip comments in extended format
                        continue
                    else:
                        # Add as custom step
                        extended_steps.append(step)

                # Create extended format YAML structure
                yaml_data = {
                    "name": f"{test_case['id']} - {test_case['description']}",
                    "description": f"{test_case['description']} (Module: {module_name}, Suite: {suite['name']})",
                    "project": "D365 Test Automation",
                    "goal": f"{module_name} Tests",
                    "journey": suite["name"],
                    "infrastructure": {
                        "starting_url": "https://[instance].crm.dynamics.com"
                    },
                    "steps": extended_steps,
                }

                # Add variables if any
                if test_data.get("variables"):
                    yaml_data["data"] = test_data["variables"]

                # Write YAML file
                with open(yaml_file_path, "w", encoding="utf-8") as f:
                    yaml.dump(
                        yaml_data,
                        f,
                        default_flow_style=False,
                        sort_keys=False,
                        allow_unicode=True,
                        width=1000,
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

        return all_yaml_files


def main():
    """Main conversion function"""
    converter = D365TestConverter()

    input_dir = "/Users/marklovelady/Downloads/D365"
    output_dir = "d365-virtuoso-tests"

    print("Starting D365 test conversion...")
    yaml_files = converter.convert_all_files(input_dir, output_dir)

    print(f"\nConversion complete! Created {len(yaml_files)} YAML files.")
    print("\nNext steps:")
    print("1. Review the generated YAML files in the d365-virtuoso-tests directory")
    print("2. Update instance URLs and credentials in the YAML files")
    print("3. Deploy tests using: ./bin/api-cli run-test <yaml-file>")


if __name__ == "__main__":
    main()
