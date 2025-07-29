/**
 * Creating a sidebar enables you to:
 - create an ordered group of docs
 - render a sidebar for each doc of that group
 - provide next/previous navigation

 The sidebars can be generated from the filesystem, or explicitly defined here.

 Create as many sidebars as you want.
 */

// @ts-check

/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
  // Main documentation sidebar
  tutorialSidebar: [
    {
      type: "category",
      label: "Getting Started",
      collapsed: false,
      items: ["intro", "quick-start", "installation", "configuration"],
    },
    {
      type: "category",
      label: "Developer Guide",
      items: [
        "developer-guide/overview",
        "developer-guide/supported-formats",
        "developer-guide/conversion-examples",
        "developer-guide/integration-patterns",
        "developer-guide/sdk-development",
        "developer-guide/best-practices",
      ],
    },
    {
      type: "category",
      label: "Test Formats",
      items: [
        "formats/selenium",
        "formats/cypress",
        "formats/playwright",
        "formats/yaml-format",
        "formats/json-format",
      ],
    },
    {
      type: "category",
      label: "Pattern Library",
      items: [
        "patterns/overview",
        {
          type: "category",
          label: "Navigation Patterns",
          items: [
            "patterns/navigation/basic-navigation",
            "patterns/navigation/scroll-patterns",
            "patterns/navigation/window-management",
          ],
        },
        {
          type: "category",
          label: "Interaction Patterns",
          items: [
            "patterns/interaction/click-patterns",
            "patterns/interaction/form-filling",
            "patterns/interaction/keyboard-actions",
            "patterns/interaction/mouse-actions",
          ],
        },
        {
          type: "category",
          label: "Assertion Patterns",
          items: [
            "patterns/assertion/element-assertions",
            "patterns/assertion/text-assertions",
            "patterns/assertion/value-assertions",
            "patterns/assertion/custom-assertions",
          ],
        },
        {
          type: "category",
          label: "Data Patterns",
          items: [
            "patterns/data/variable-storage",
            "patterns/data/cookie-management",
            "patterns/data/api-integration",
          ],
        },
      ],
    },
    {
      type: "category",
      label: "Troubleshooting",
      items: [
        "troubleshooting/common-errors",
        "troubleshooting/import-errors",
        "troubleshooting/timeout-issues",
        "troubleshooting/auth-failures",
        "troubleshooting/conversion-failures",
        "troubleshooting/debugging-techniques",
        "troubleshooting/performance-optimization",
      ],
    },
    {
      type: "category",
      label: "Architecture",
      items: [
        "architecture/system-design",
        "architecture/data-flow",
        "architecture/security",
        "architecture/scaling",
        "architecture/disaster-recovery",
        "architecture/deployment-strategies",
      ],
    },
    {
      type: "category",
      label: "Advanced Topics",
      items: [
        "advanced/custom-patterns",
        "advanced/webhook-integration",
        "advanced/batch-processing",
        "advanced/monitoring-logging",
        "advanced/rate-limiting",
        "advanced/caching-strategies",
      ],
    },
  ],
};

module.exports = sidebars;
