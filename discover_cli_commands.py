#!/usr/bin/env python3
# discover_cli_commands.py
import subprocess
import json
from datetime import datetime


def discover_cli_commands():
    """Discover all available CLI commands"""
    commands = {"basic": [], "advanced": [], "unsupported": []}

    # Test commands we expect to exist based on Selenium pattern mapping
    test_commands = [
        "step-navigate",
        "step-click",
        "step-fill",
        "step-clear",
        "step-hover",
        "step-double-click",
        "step-right-click",
        "step-drag-drop",
        "step-screenshot",
        "step-select",
        "step-switch-frame",
        "step-switch-window",
        "step-alert",
        "step-cookie",
        "step-assert",
        "step-keyboard",
        "step-wait",
        "step-scroll",
        "step-refresh",
        "step-back",
        "step-interact",
        "step-data",
        "step-dialog",
        "step-file",
        "step-misc",
        "step-window",
    ]

    # Use the local CLI binary
    cli_path = "./bin/api-cli"

    for cmd in test_commands:
        try:
            result = subprocess.run(
                [cli_path, cmd, "--help"], capture_output=True, text=True, timeout=5
            )

            if result.returncode == 0:
                commands["basic"].append(cmd)
                print(f"✅ {cmd}: Supported")
            else:
                commands["unsupported"].append(cmd)
                print(f"❌ {cmd}: Not found")

        except Exception as e:
            commands["unsupported"].append(cmd)
            print(f"❌ {cmd}: Error - {e}")

    # Also check for subcommands in step-interact, step-assert, etc.
    check_subcommands = {
        "step-interact": [
            "click",
            "double-click",
            "right-click",
            "hover",
            "drag-drop",
            "type",
            "clear",
            "select",
        ],
        "step-assert": ["text", "visible", "enabled", "attribute", "css", "count"],
        "step-navigate": ["to", "back", "forward", "refresh", "scroll"],
        "step-wait": ["for-element", "for-text", "for-time", "for-condition"],
        "step-window": ["switch", "close", "maximize", "resize"],
        "step-dialog": ["accept", "dismiss", "send-text"],
        "step-data": ["store", "cookie", "local-storage", "session-storage"],
    }

    print("\n=== CHECKING SUBCOMMANDS ===")
    subcommand_support = {}

    for parent_cmd, subcommands in check_subcommands.items():
        if parent_cmd in commands["basic"]:
            subcommand_support[parent_cmd] = []
            for subcmd in subcommands:
                try:
                    result = subprocess.run(
                        [cli_path, parent_cmd, subcmd, "--help"],
                        capture_output=True,
                        text=True,
                        timeout=5,
                    )

                    if result.returncode == 0:
                        subcommand_support[parent_cmd].append(subcmd)
                        print(f"✅ {parent_cmd} {subcmd}: Supported")
                    else:
                        print(f"❌ {parent_cmd} {subcmd}: Not found")
                except:
                    print(f"❌ {parent_cmd} {subcmd}: Error")

    return commands, subcommand_support


if __name__ == "__main__":
    commands, subcommands = discover_cli_commands()

    print("\n=== CLI COMMAND SUMMARY ===")
    print(f"Supported: {len(commands['basic'])}")
    print(f"Unsupported: {len(commands['unsupported'])}")

    report = {
        "timestamp": datetime.now().isoformat(),
        "working_commands": commands["basic"],
        "unsupported_commands": commands["unsupported"],
        "subcommand_support": subcommands,
    }

    with open("cli_commands_report.json", "w") as f:
        json.dump(report, f, indent=2)
