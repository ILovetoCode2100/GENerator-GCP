// @ts-check
// Note: type annotations allow type checking and IDEs autocompletion

const lightCodeTheme = require("prism-react-renderer/themes/github");
const darkCodeTheme = require("prism-react-renderer/themes/dracula");

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: "Virtuoso Test Converter API",
  tagline:
    "Convert and run tests from Selenium, Cypress, and Playwright to Virtuoso format",
  favicon: "img/favicon.ico",

  // Production URL
  url: "https://virtuoso-converter.docs.example.com",
  baseUrl: "/",

  // GitHub pages deployment config
  organizationName: "virtuoso",
  projectName: "virtuoso-generator",

  onBrokenLinks: "throw",
  onBrokenMarkdownLinks: "warn",

  i18n: {
    defaultLocale: "en",
    locales: ["en"],
  },

  presets: [
    [
      "classic",
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          sidebarPath: require.resolve("./sidebars.js"),
          editUrl:
            "https://github.com/virtuoso/virtuoso-generator/tree/main/documentation/",
          showLastUpdateAuthor: true,
          showLastUpdateTime: true,
          remarkPlugins: [],
          rehypePlugins: [],
        },
        blog: false,
        theme: {
          customCss: require.resolve("./src/css/custom.css"),
        },
      }),
    ],
  ],

  plugins: [
    [
      "@docusaurus/plugin-content-docs",
      {
        id: "api",
        path: "api",
        routeBasePath: "api",
        sidebarPath: require.resolve("./sidebars-api.js"),
      },
    ],
    [
      "@docusaurus/plugin-ideal-image",
      {
        quality: 70,
        max: 1030,
        min: 640,
        steps: 2,
        disableInDev: false,
      },
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      // Social card
      image: "img/virtuoso-social-card.jpg",

      navbar: {
        title: "Virtuoso Test Converter",
        logo: {
          alt: "Virtuoso Logo",
          src: "img/logo.svg",
        },
        items: [
          {
            type: "docSidebar",
            sidebarId: "tutorialSidebar",
            position: "left",
            label: "Documentation",
          },
          {
            to: "/api/overview",
            label: "API Reference",
            position: "left",
          },
          {
            to: "/docs/patterns",
            label: "Pattern Library",
            position: "left",
          },
          {
            href: "https://github.com/virtuoso/virtuoso-generator",
            label: "GitHub",
            position: "right",
          },
        ],
      },

      footer: {
        style: "dark",
        links: [
          {
            title: "Documentation",
            items: [
              {
                label: "Quick Start",
                to: "/docs/intro",
              },
              {
                label: "API Reference",
                to: "/api/overview",
              },
              {
                label: "Developer Guide",
                to: "/docs/developer-guide",
              },
            ],
          },
          {
            title: "Resources",
            items: [
              {
                label: "Pattern Library",
                to: "/docs/patterns",
              },
              {
                label: "Troubleshooting",
                to: "/docs/troubleshooting",
              },
              {
                label: "Architecture",
                to: "/docs/architecture",
              },
            ],
          },
          {
            title: "Community",
            items: [
              {
                label: "Stack Overflow",
                href: "https://stackoverflow.com/questions/tagged/virtuoso",
              },
              {
                label: "Discord",
                href: "https://discord.gg/virtuoso",
              },
              {
                label: "GitHub",
                href: "https://github.com/virtuoso/virtuoso-generator",
              },
            ],
          },
        ],
        copyright: `Copyright © ${new Date().getFullYear()} Virtuoso. Built with Docusaurus.`,
      },

      prism: {
        theme: lightCodeTheme,
        darkTheme: darkCodeTheme,
        additionalLanguages: ["bash", "yaml", "json", "go", "python"],
      },

      algolia: {
        appId: "YOUR_APP_ID",
        apiKey: "YOUR_API_KEY",
        indexName: "virtuoso-converter",
        contextualSearch: true,
      },

      announcementBar: {
        id: "support_us",
        content:
          '⭐ If you like Virtuoso Test Converter, give it a star on <a target="_blank" rel="noopener noreferrer" href="https://github.com/virtuoso/virtuoso-generator">GitHub</a>!',
        backgroundColor: "#fafbfc",
        textColor: "#091E42",
        isCloseable: false,
      },
    }),
};

module.exports = config;
