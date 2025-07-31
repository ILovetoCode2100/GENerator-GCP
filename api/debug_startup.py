#!/usr/bin/env python
"""Debug startup issues"""

import sys
import os

print(f"Python version: {sys.version}")
print(f"Current directory: {os.getcwd()}")
print(f"Python path: {sys.path}")

print("\nDirectory structure:")
for root, dirs, files in os.walk("/app"):
    level = root.replace("/app", "").count(os.sep)
    indent = " " * 2 * level
    print(f"{indent}{os.path.basename(root)}/")
    subindent = " " * 2 * (level + 1)
    for file in files:
        if file.endswith('.py'):
            print(f"{subindent}{file}")

print("\n\nChecking if tests.py exists:")
import os
tests_path = "/app/app/routes/tests.py"
if os.path.exists(tests_path):
    print(f"✓ {tests_path} exists")
    print(f"  File size: {os.path.getsize(tests_path)} bytes")
else:
    print(f"✗ {tests_path} does NOT exist")

print("\n\nTesting imports:")
try:
    import app.routes.tests
    print("✓ Successfully imported app.routes.tests")
except Exception as e:
    print(f"✗ Failed to import app.routes.tests: {e}")
    import traceback
    traceback.print_exc()

print("\n\nChecking sys.modules:")
import sys
if 'app' in sys.modules:
    print("✓ 'app' is in sys.modules")
if 'app.routes' in sys.modules:
    print("✓ 'app.routes' is in sys.modules")

print("\n\nTrying to start the app:")
try:
    from app.main import app
    print("✓ Successfully imported app")
except Exception as e:
    print(f"✗ Failed to import app: {e}")
    import traceback
    traceback.print_exc()