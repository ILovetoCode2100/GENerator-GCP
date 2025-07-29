/**
 * API Reference Sidebar Configuration
 */

// @ts-check

/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebarsApi = {
  apiSidebar: [
    "overview",
    "authentication",
    "rate-limiting",
    "error-handling",
    {
      type: "category",
      label: "Core Endpoints",
      collapsed: false,
      items: [
        {
          type: "category",
          label: "Convert",
          items: [
            "endpoints/convert/overview",
            "endpoints/convert/selenium-to-virtuoso",
            "endpoints/convert/cypress-to-virtuoso",
            "endpoints/convert/playwright-to-virtuoso",
            "endpoints/convert/batch-conversion",
          ],
        },
        {
          type: "category",
          label: "Status",
          items: [
            "endpoints/status/job-status",
            "endpoints/status/list-jobs",
            "endpoints/status/cancel-job",
          ],
        },
        {
          type: "category",
          label: "Patterns",
          items: [
            "endpoints/patterns/list-patterns",
            "endpoints/patterns/pattern-details",
            "endpoints/patterns/confidence-scores",
            "endpoints/patterns/custom-patterns",
          ],
        },
        {
          type: "category",
          label: "Feedback",
          items: [
            "endpoints/feedback/submit-feedback",
            "endpoints/feedback/improvement-suggestions",
            "endpoints/feedback/pattern-training",
          ],
        },
      ],
    },
    {
      type: "category",
      label: "Command API",
      items: [
        "commands/overview",
        "commands/execute",
        "commands/list-commands",
        "commands/command-help",
        "commands/batch-execution",
      ],
    },
    {
      type: "category",
      label: "Test Management",
      items: [
        "tests/run-test",
        "tests/upload-test",
        "tests/test-templates",
        "tests/test-results",
        "tests/test-history",
      ],
    },
    {
      type: "category",
      label: "Session Management",
      items: [
        "sessions/create-session",
        "sessions/list-sessions",
        "sessions/get-session",
        "sessions/update-session",
        "sessions/delete-session",
        "sessions/activate-session",
      ],
    },
    {
      type: "category",
      label: "Webhooks",
      items: [
        "webhooks/overview",
        "webhooks/webhook-events",
        "webhooks/webhook-security",
        "webhooks/webhook-examples",
      ],
    },
    {
      type: "category",
      label: "SDKs",
      items: [
        "sdks/overview",
        "sdks/javascript",
        "sdks/python",
        "sdks/java",
        "sdks/csharp",
        "sdks/go",
      ],
    },
    {
      type: "category",
      label: "OpenAPI",
      items: [
        "openapi/specification",
        "openapi/schema-definitions",
        "openapi/code-generation",
        "openapi/postman-collection",
      ],
    },
  ],
};

module.exports = sidebarsApi;
