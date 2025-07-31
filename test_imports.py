#!/usr/bin/env python3
"""Test import structure"""

import sys
import os

# Add the api directory to Python path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), 'api'))

print("Testing imports...")

try:
    print("1. Importing app...")
    import app
    print("   ✓ app imported successfully")
except Exception as e:
    print(f"   ✗ Failed to import app: {e}")

try:
    print("2. Importing app.routes...")
    import app.routes
    print("   ✓ app.routes imported successfully")
except Exception as e:
    print(f"   ✗ Failed to import app.routes: {e}")

try:
    print("3. Importing app.routes.tests...")
    import app.routes.tests
    print("   ✓ app.routes.tests imported successfully")
except Exception as e:
    print(f"   ✗ Failed to import app.routes.tests: {e}")

try:
    print("4. Importing from app.routes...")
    from app.routes import tests
    print("   ✓ from app.routes import tests succeeded")
except Exception as e:
    print(f"   ✗ Failed to import from app.routes: {e}")

print("\nChecking module locations:")
if 'app' in sys.modules:
    print(f"app module location: {sys.modules['app'].__file__ if hasattr(sys.modules['app'], '__file__') else 'No __file__ attribute'}")
if 'app.routes' in sys.modules:
    print(f"app.routes module location: {sys.modules['app.routes'].__file__ if hasattr(sys.modules['app.routes'], '__file__') else 'No __file__ attribute'}")
if 'app.routes.tests' in sys.modules:
    print(f"app.routes.tests module location: {sys.modules['app.routes.tests'].__file__ if hasattr(sys.modules['app.routes.tests'], '__file__') else 'No __file__ attribute'}")