import {
  allClusterNames,
  allClusterIds,
  clusterIdByName,
} from "../fixtures/clusters-overview/available_clusters";

context("Clusters Overview", () => {
  const availableClusters = allClusterNames();
  before(() => {
    cy.resetDatabase();
    cy.loadScenario("healthy-27-node-SAP-cluster");
    cy.loadChecksCatalog("checks-catalog/catalog.json");
    cy.loadChecksResults(
      "clusters-overview/checks_results_critical.json",
      "04b8f8c21f9fd8991224478e8c4362f8"
    );
    cy.loadChecksResults(
      "clusters-overview/checks_results_warning.json",
      "4e905d706da85f5be14f85fa947c1e39"
    );
    cy.loadChecksResults(
      "clusters-overview/checks_results_passing.json",
      "9c832998801e28cd70ad77380e82a5c0"
    );

    cy.visit("/");
    cy.navigateToItem("Clusters");
    cy.url().should("include", "/clusters");
  });

  describe("Registered Clusters should be available in the overview", () => {
    it("should show 9 of the 9 registered clusters with default pagination settings", () => {
      cy.get(".tn-clustername").its("length").should("eq", 9);
    });
    it("should show 9 as total items in the pagination controls", () => {
      cy.get(".pagination-count").should("contain", "9 items");
    });
    it("should have 1 pages", () => {
      cy.get(".page-item").its("length").should("eq", 3); // We add +2 to the page count because of the first and last page
    });
    describe("Discovered clusternames are the expected ones", () => {
      availableClusters.forEach((clusterName) => {
        it(`should have a cluster named ${clusterName}`, () => {
          cy.get(".tn-clustername").each(($link) => {
            const displayedClusterName = $link.text().trim();
            expect(availableClusters).to.include(displayedClusterName);
          });
        });
      });
    });
  });

  describe("Health Detection", () => {
    describe("Health Container shows the health overview of all Clusters", () => {
      it("should show health status of the entire cluster of 9 hosts with partial pagination", () => {
        cy.reloadList("clusters", 10);
        cy.get(".health-container .health-passing").should("contain", 1);
      });
      it("should show health status of all 9 clusters", () => {
        cy.reloadList("clusters", 100);
        cy.get(".health-container .health-passing").should("contain", 1);
        cy.get(".health-container .health-warning").should("contain", 1);
        cy.get(".health-container .health-critical").should("contain", 1);
      });
    });
  });

  describe("Clusters Tagging", () => {
    before(() => {
      cy.get("body").then(($body) => {
        const deleteTag = ".tn-cluster-tags x";
        if ($body.find(deleteTag).length > 0) {
          cy.get(deleteTag).then(($deleteTag) =>
            cy.wrap($deleteTag).click({ multiple: true })
          );
        }
      });
    });
    const clustersByMatchingPattern = (pattern) => (clusterName) =>
      clusterName.includes(pattern);
    const taggingRules = [
      ["hana_cluster_1", "env1"],
      ["hana_cluster_2", "env2"],
      ["hana_cluster_3", "env3"],
    ];

    taggingRules.forEach(([pattern, tag]) => {
      describe(`Add tag '${tag}' to all clusters with '${pattern}' in the cluster name`, () => {
        availableClusters
          .filter(clustersByMatchingPattern(pattern))
          .forEach((clusterName) => {
            it(`should tag cluster '${clusterName}'`, () => {
              cy.get(
                `#cluster-${clusterIdByName(
                  clusterName
                )} > .tn-clusters-tags > .tagify`
              )
                .type(tag)
                .trigger("change");
            });
          });
      });
    });
  });

  describe("Filtering the Clusters overview", () => {
    before(() => {
      cy.reloadList("clusters", 100);
    });

    const resetFilter = (option) => {
      cy.intercept("GET", `/clusters?per_page=100`).as("resetFilter");
      cy.get(option).click();
      cy.wait("@resetFilter");
    };

    describe("Filtering by health", () => {
      before(() => {
        cy.get(".tn-filters > :nth-child(2) > .btn").click();
      });
      const healthScenarios = [
        ["passing", 1],
        ["warning", 1],
        ["critical", 1],
      ];
      healthScenarios.forEach(
        ([health, expectedClustersWithThisHealth], index) => {
          it(`should show ${
            expectedClustersWithThisHealth || "an empty list of"
          } clusters when filtering by health '${health}'`, () => {
            cy.intercept("GET", `/clusters?per_page=100&health=${health}`).as(
              "filterByHealthStatus"
            );
            const selectedOption = `#bs-select-1-${index}`;
            cy.get(selectedOption).click();
            cy.wait("@filterByHealthStatus").then(() => {
              expectedClustersWithThisHealth == 0 &&
                cy
                  .get(".table.eos-table")
                  .contains("There are currently no records to be shown");
              expectedClustersWithThisHealth > 0 &&
                cy
                  .get(".tn-clustername")
                  .its("length")
                  .should("eq", expectedClustersWithThisHealth);
              cy.get(".pagination-count").should(
                "contain",
                `${expectedClustersWithThisHealth} items`
              );
              cy.get(".page-item")
                .its("length")
                .should(
                  "eq",
                  Math.ceil(expectedClustersWithThisHealth / 100) + 2
                );
              resetFilter(selectedOption);
            });
          });
        }
      );
    });

    describe("Filtering by SAP system", () => {
      before(() => {
        cy.get(".tn-filters > :nth-child(4) > .btn").click();
      });
      const SAPSystemsScenarios = [
        ["HDD", 1],
        ["HDP", 1],
        ["HDQ", 1],
      ];
      SAPSystemsScenarios.forEach(
        ([sapsystem, expectedRelatedClusters], index) => {
          it(`should have ${expectedRelatedClusters} clusters related to SAP system '${sapsystem}'`, () => {
            cy.intercept("GET", `/clusters?per_page=100&sids=${sapsystem}`).as(
              "filterBySAPSystem"
            );
            const selectedOption = `#bs-select-3-${index}`;
            cy.get(selectedOption).click();
            cy.wait("@filterBySAPSystem").then(() => {
              cy.get(".tn-clustername")
                .its("length")
                .should("eq", expectedRelatedClusters);
              cy.get(".pagination-count").should(
                "contain",
                `${expectedRelatedClusters} items`
              );
              cy.get(".page-item")
                .its("length")
                .should("eq", Math.ceil(expectedRelatedClusters / 100) + 2);
            });
            resetFilter(selectedOption);
          });
        }
      );
    });

    describe("Filtering by tags", () => {
      before(() => {
        cy.get(".tn-filters > :nth-child(5) > .btn").click();
      });
      const tagsScenarios = [
        ["env1", 1],
        ["env2", 1],
        ["env3", 1],
      ];
      tagsScenarios.forEach(([tag, expectedTaggedClusters], index) => {
        it(`should have ${expectedTaggedClusters} clusters tagged with tag '${tag}'`, () => {
          cy.intercept("GET", `/clusters?per_page=100&tags=${tag}`).as(
            "filterByTags"
          );
          const selectedOption = `#bs-select-3-${index}`;
          cy.get(selectedOption).click();
          cy.wait("@filterByTags").then(() => {
            cy.get(".tn-clustername")
              .its("length")
              .should("eq", expectedTaggedClusters);
            cy.get(".pagination-count").should(
              "contain",
              `${expectedTaggedClusters} items`
            );
            cy.get(".page-item")
              .its("length")
              .should("eq", Math.ceil(expectedTaggedClusters / 100) + 2);
            resetFilter(selectedOption);
          });
        });
      });
    });

    describe("Removing filtered tags", () => {
      const tag = "tag1";
      before(() => {
        cy.intercept("/api/clusters/**").as("tagPosted");
        cy.intercept("GET", "/api/tags?resource_type=clusters").as(
          "filterRefreshed"
        );
        cy.get(`#cluster-${availableClusters[0]} > .tn-cluster-tags > .tagify`)
          .type(tag)
          .click();
        cy.wait("@tagPosted");
        cy.wait("@filterRefreshed");
        cy.get(`#cluster-${availableClusters[1]} > .tn-cluster-tags > .tagify`)
          .type(tag)
          .click();
        cy.wait("@tagPosted");
        cy.wait("@filterRefreshed");
        cy.get(".dropdown-item").contains(tag);
        cy.get(".tn-filters > :nth-child(4) > .btn").click();
      });

      it(`should reload the clusters table when filtered tags are removed`, () => {
        cy.intercept("GET", `/clusters?per_page=100&tags=${tag}`).as(
          "filterByTags"
        );
        cy.get(".dropdown-item").contains(tag).click();
        cy.wait("@filterByTags").then(() => {
          cy.get(".tn-clustername").should("have.length", 2);
        });

        cy.get(".tn-filters > :nth-child(4) > .btn").click();

        cy.intercept("GET", `/clusters?per_page=100&tags=${tag}`).as(
          "firstTagRemoved"
        );
        cy.get(`#cluster-${availableClusters[0]} > .tn-cluster-tags tag`)
          .filter(`[value="${tag}"]`)
          .find("> x")
          .click();
        cy.wait("@firstTagRemoved").then(() => {
          cy.get(".tn-clustername").should("have.length", 1);
        });

        cy.intercept("GET", "/clusters?per_page=100").as("secondTagRemoved");
        cy.get(`#cluster-${availableClusters[1]} > .tn-cluster-tags tag`)
          .filter(`[value="${tag}"]`)
          .find("> x")
          .click();
        cy.wait("@secondTagRemoved").then(() => {
          cy.get(".dropdown-item").contains(tag).should("not.exist");
          cy.get(".tn-clustername").should(
            "have.length",
            availableClusters.length
          );
        });
      });
    });
  });
});
