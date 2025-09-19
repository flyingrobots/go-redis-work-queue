---
noteID: f1dd9a08-af1a-4e6b-9e8e-9b370c802d3e
---
# CodeRabbit Prompt Worksheet

Source: coderabbit-pr3-prompts-latest.md

## Progress

`██████████░░░░░░░░░░`

Addressed 284 / 586 (48.5%)

## Table of Contents

- [x] [In .github/CODEOWNERS around lines 5 to 18, the file duplicates ownership](#in-githubcodeowners-around-lines-5-to-18-the-file-duplicates-ownership)
- [x] [</summary>](#summary)
- [x] [.github/workflows/changelog.yml around lines 28-29: the command uses unquoted](#githubworkflowschangelogyml-around-lines-28-29-the-command-uses-unquoted)
- [ ] [In .github/workflows/changelog.yml around lines 36-38, the workflow pushes](#in-githubworkflowschangelogyml-around-lines-36-38-the-workflow-pushes)
- [x] [.github/workflows/ci.yml around lines 38 to 45: the workflow immediately starts](#githubworkflowsciyml-around-lines-38-to-45-the-workflow-immediately-starts)
- [x] [In .github/workflows/goreleaser.yml around lines 25 to 27, the echo redirections](#in-githubworkflowsgoreleaseryml-around-lines-25-to-27-the-echo-redirections)
- [ ] [In .github/workflows/goreleaser.yml around lines 28-33, the workflow logs into](#in-githubworkflowsgoreleaseryml-around-lines-28-33-the-workflow-logs-into)
- [x] [.github/workflows/markdownlint.yml around line 6: the branch array uses spaces](#githubworkflowsmarkdownlintyml-around-line-6-the-branch-array-uses-spaces)
- [x] [.github/workflows/markdownlint.yml lines 12-21: the workflow lacks](#githubworkflowsmarkdownlintyml-lines-12-21-the-workflow-lacks)
- [x] [.github/workflows/markdownlint.yml lines 12–16: the workflow uses mutable tags](#githubworkflowsmarkdownlintyml-lines-1216-the-workflow-uses-mutable-tags)
- [x] [.goreleaser.yaml around lines 15 to 20: the archives block currently produces](#goreleaseryaml-around-lines-15-to-20-the-archives-block-currently-produces)
- [x] [.goreleaser.yaml around lines 38 to 41: the current owner/name fields use bare](#goreleaseryaml-around-lines-38-to-41-the-current-ownername-fields-use-bare)
- [x] [.goreleaser.yaml around lines 43 to 45: there is an extra trailing blank line](#goreleaseryaml-around-lines-43-to-45-there-is-an-extra-trailing-blank-line)
- [x] [In .markdownlint.yaml at line 4 the config disables MD013 repo-wide; remove the](#in-markdownlintyaml-at-line-4-the-config-disables-md013-repo-wide-remove-the)
- [x] [.vscode/extensions.json lines 1-6: The workspace recommendations only include](#vscodeextensionsjson-lines-1-6-the-workspace-recommendations-only-include)
- [x] [In .vscode/settings.json around lines 9 to 13, the workspace is not enabling](#in-vscodesettingsjson-around-lines-9-to-13-the-workspace-is-not-enabling)
- [x] [In CHANGELOG.md around lines 17 to 30, the current entry is a freeform](#in-changelogmd-around-lines-17-to-30-the-current-entry-is-a-freeform)
- [x] [In create_review_tasks.py around lines 22-24 (and also update the similar logic](#in-createreviewtaskspy-around-lines-22-24-and-also-update-the-similar-logic)
- [x] [In demos/responsive-tui.tape around lines 20 to 27 (and similarly at 81-88 and](#in-demosresponsive-tuitape-around-lines-20-to-27-and-similarly-at-81-88-and)
- [x] [In dependency_analysis.py around lines 233 to 243, the infrastructure dict is](#in-dependencyanalysispy-around-lines-233-to-243-the-infrastructure-dict-is)
- [x] [In deploy/docker-compose.yml around lines 19 to 22 the service volume mounts](#in-deploydocker-composeyml-around-lines-19-to-22-the-service-volume-mounts)
- [x] [In deploy/grafana/dashboards/work-queue.json around lines 6 to 9, the PromQL](#in-deploygrafanadashboardswork-queuejson-around-lines-6-to-9-the-promql)
- [x] [In deploy/grafana/dashboards/work-queue.json around lines 20-24 (and likewise](#in-deploygrafanadashboardswork-queuejson-around-lines-20-24-and-likewise)
- [x] [In deploy/grafana/dashboards/work-queue.json around lines 26 to 29, the](#in-deploygrafanadashboardswork-queuejson-around-lines-26-to-29-the)
- [x] [deployments/admin-api/k8s-redis.yaml lines 1-62: the manifest lacks any](#deploymentsadmin-apik8s-redisyaml-lines-1-62-the-manifest-lacks-any)
- [x] [In deployments/admin-api/k8s-redis.yaml around lines 62 to 74, the Service](#in-deploymentsadmin-apik8s-redisyaml-around-lines-62-to-74-the-service)
- [x] [In docs/01_product_roadmap.md around lines 34 to 39, the roadmap dates list RC](#in-docs01productroadmapmd-around-lines-34-to-39-the-roadmap-dates-list-rc)
- [x] [In docs/01_product_roadmap.md around lines 48–52, the "Dependencies" bullets are](#in-docs01productroadmapmd-around-lines-4852-the-dependencies-bullets-are)
- [x] [In docs/03_milestones.md around lines 6 to 8, the milestone entries lack](#in-docs03milestonesmd-around-lines-6-to-8-the-milestone-entries-lack)
- [x] [In docs/04_sprint_plans.md around lines 5 to 8, replace the ambiguous term](#in-docs04sprintplansmd-around-lines-5-to-8-replace-the-ambiguous-term)
- [x] [In docs/06_technical_spec.md around lines 113–117, the reaper section is](#in-docs06technicalspecmd-around-lines-113117-the-reaper-section-is)
- [x] [In docs/13_release_versioning.md around lines 21 to 25, the release checklist](#in-docs13releaseversioningmd-around-lines-21-to-25-the-release-checklist)
- [x] [In docs/13_release_versioning.md around lines 26 to 31, the current instructions](#in-docs13releaseversioningmd-around-lines-26-to-31-the-current-instructions)
- [x] [In docs/YOU ARE WORKER 6.rb around lines 3 to 4, the README uses vendor-specific](#in-docsyou-are-worker-6rb-around-lines-3-to-4-the-readme-uses-vendor-specific)
- [x] [In docs/YOU ARE WORKER 6.rb around lines 6 to 13, the workflow step that tells](#in-docsyou-are-worker-6rb-around-lines-6-to-13-the-workflow-step-that-tells)
- [x] [In docs/YOU ARE WORKER 6.rb around lines 21 to 26, the example shell commands](#in-docsyou-are-worker-6rb-around-lines-21-to-26-the-example-shell-commands)
- [x] [In auto_commit.sh around line 1, the script lacks Bash strict mode which can](#in-autocommitsh-around-line-1-the-script-lacks-bash-strict-mode-which-can)
- [x] [In auto_commit.sh around lines 4-6 (and similarly lines 45-47) the script starts](#in-autocommitsh-around-lines-4-6-and-similarly-lines-45-47-the-script-starts)
- [x] [In auto_commit.sh around lines 30 to 41, the script currently parses git push](#in-autocommitsh-around-lines-30-to-41-the-script-currently-parses-git-push)
- [x] [In config/config.example.yaml around line 2, the default Redis address is set to](#in-configconfigexampleyaml-around-line-2-the-default-redis-address-is-set-to)
- [x] [In config/config.example.yaml around lines 50 to 66, the idempotency settings](#in-configconfigexampleyaml-around-lines-50-to-66-the-idempotency-settings)
- [x] [In config/config.example.yaml around lines 67 to 80, the outbox section is](#in-configconfigexampleyaml-around-lines-67-to-80-the-outbox-section-is)
- [x] [In create_review_tasks.py around lines 1-4 and also update lines 102-112,](#in-createreviewtaskspy-around-lines-1-4-and-also-update-lines-102-112)
- [x] [In create_review_tasks.py around lines 9 to 11, the code calls](#in-createreviewtaskspy-around-lines-9-to-11-the-code-calls)
- [x] [In create_review_tasks.py around lines 30-31, the code constructs timestamps](#in-createreviewtaskspy-around-lines-30-31-the-code-constructs-timestamps)
- [x] [In demos/lipgloss-transformation.tape around lines 136-142 (and similarly for](#in-demoslipgloss-transformationtape-around-lines-136-142-and-similarly-for)
- [x] [In demos/lipgloss-transformation.tape around lines 271 to 276, the script calls](#in-demoslipgloss-transformationtape-around-lines-271-to-276-the-script-calls)
- [ ] [In demos/responsive-tui.tape around lines 271-278 the final figlet call can](#in-demosresponsive-tuitape-around-lines-271-278-the-final-figlet-call-can)
- [x] [In dependency_analysis.py around lines 7 to 231, feature keys use kebab-case](#in-dependencyanalysispy-around-lines-7-to-231-feature-keys-use-kebab-case)
- [x] [In docs/05_architecture.md around line 7, the architecture text omits the](#in-docs05architecturemd-around-line-7-the-architecture-text-omits-the)
- [x] [In docs/06_technical_spec.md around lines 124-129 and also 131-134, the metrics](#in-docs06technicalspecmd-around-lines-124-129-and-also-131-134-the-metrics)
- [x] [In docs/09_requirements.md around lines 43-49, the acceptance criteria are](#in-docs09requirementsmd-around-lines-43-49-the-acceptance-criteria-are)
- [x] [In docs/12_performance_baseline.md around lines 26 to 28, the example starts a](#in-docs12performancebaselinemd-around-lines-26-to-28-the-example-starts-a)
- [x] [In docs/12_performance_baseline.md around lines 31 to 33, the README tells users](#in-docs12performancebaselinemd-around-lines-31-to-33-the-readme-tells-users)
- [x] [In docs/14_ops_runbook.md around lines 21 to 26, replace the single unpinned](#in-docs14opsrunbookmd-around-lines-21-to-26-replace-the-single-unpinned)
- [x] [In docs/14_ops_runbook.md around lines 32 to 35, the env-var mapping description](#in-docs14opsrunbookmd-around-lines-32-to-35-the-env-var-mapping-description)
- [x] [In docs/14_ops_runbook.md around lines 38 to 42, the guidance currently](#in-docs14opsrunbookmd-around-lines-38-to-42-the-guidance-currently)
- [x] [In docs/14_ops_runbook.md around lines 51 to 75, the purge-dlq example and](#in-docs14opsrunbookmd-around-lines-51-to-75-the-purge-dlq-example-and)
- [x] [In docs/api/dlq-remediation-pipeline.md around lines 121 to 137, the sample](#in-docsapidlq-remediation-pipelinemd-around-lines-121-to-137-the-sample)
- [ ] [In docs/api/dlq-remediation-pipeline.md around lines 197 to 252, the matcher](#in-docsapidlq-remediation-pipelinemd-around-lines-197-to-252-the-matcher)
- [x] [In docs/api/dlq-remediation-pipeline.md around lines 541–606, the audit example](#in-docsapidlq-remediation-pipelinemd-around-lines-541606-the-audit-example)
- [x] [In docs/api/dlq-remediation-pipeline.md around lines 860 to 876, the WebSocket](#in-docsapidlq-remediation-pipelinemd-around-lines-860-to-876-the-websocket)
- [x] [In docs/api/dlq-remediation-ui.md around line 9, the doc currently states the](#in-docsapidlq-remediation-uimd-around-line-9-the-doc-currently-states-the)
- [x] [In docs/api/dlq-remediation-ui.md around lines 231 to 241, the API currently](#in-docsapidlq-remediation-uimd-around-lines-231-to-241-the-api-currently)
- [x] [In docs/api/dlq-remediation-ui.md around lines 387 to 394, the documented rate](#in-docsapidlq-remediation-uimd-around-lines-387-to-394-the-documented-rate)
- [x] [In docs/SLAPS/FINAL-POSTMORTEM.md around lines 8-9 (and also lines 324-336), the](#in-docsslapsfinal-postmortemmd-around-lines-8-9-and-also-lines-324-336-the)
- [x] [In README.md around lines 3 to 6, the intro mentions features without a clear,](#in-readmemd-around-lines-3-to-6-the-intro-mentions-features-without-a-clear)
- [ ] [In README.md around lines 38-49 there is a mismatch between the documented Go](#in-readmemd-around-lines-38-49-there-is-a-mismatch-between-the-documented-go)
- [ ] [README.md lines 123-145: the "Purge all (test keys)" admin command is presented](#readmemd-lines-123-145-the-purge-all-test-keys-admin-command-is-presented)
- [ ] [In README.md around lines 149 to 155, the docs claim metrics/health are served](#in-readmemd-around-lines-149-to-155-the-docs-claim-metricshealth-are-served)
- [ ] [In README.md around lines 166 to 185, the provided docker run example starts the](#in-readmemd-around-lines-166-to-185-the-provided-docker-run-example-starts-the)
- [x] [In append_metadata.py around line 11, the script uses a hardcoded absolute path](#in-appendmetadatapy-around-line-11-the-script-uses-a-hardcoded-absolute-path)
- [x] [In auto_commit.sh around lines 8 to 10, the current use of "ls ... | wc -l | tr](#in-autocommitsh-around-lines-8-to-10-the-current-use-of-ls---wc--l--tr)
- [x] [In auto_commit.sh around lines 16–28, the git commit call currently silences](#in-autocommitsh-around-lines-1628-the-git-commit-call-currently-silences)
- [x] [In BUGS.md around lines 27-28: the repo currently depends on both](#in-bugsmd-around-lines-27-28-the-repo-currently-depends-on-both)
- [ ] [In claude_worker.py around line 35, the return type hint uses Optional[Path] but](#in-claudeworkerpy-around-line-35-the-return-type-hint-uses-optionalpath-but)
- [x] [In claude_worker.py around lines 90-145 the worker currently blocks on](#in-claudeworkerpy-around-lines-90-145-the-worker-currently-blocks-on)
- [x] [In claude_worker.py around lines 157-159 the bare "except: pass" silently](#in-claudeworkerpy-around-lines-157-159-the-bare-except-pass-silently)
- [x] [In deployments/admin-api/monitoring.yaml lines 1-66: this ConfigMap duplicates](#in-deploymentsadmin-apimonitoringyaml-lines-1-66-this-configmap-duplicates)
- [x] [In deployments/admin-api/monitoring.yaml around line 5 (and also line 71), the](#in-deploymentsadmin-apimonitoringyaml-around-line-5-and-also-line-71-the)
- [x] [In deployments/admin-api/monitoring.yaml around lines 58-65, the alert divides](#in-deploymentsadmin-apimonitoringyaml-around-lines-58-65-the-alert-divides)
- [x] [In deployments/admin-api/monitoring.yaml around lines 82 to 99 (and also lines](#in-deploymentsadmin-apimonitoringyaml-around-lines-82-to-99-and-also-lines)
- [x] [In deployments/admin-api/monitoring.yaml around line 128, the file currently](#in-deploymentsadmin-apimonitoringyaml-around-line-128-the-file-currently)
- [x] [In deployments/kubernetes/monitoring.yaml around lines 1 to 17, the](#in-deploymentskubernetesmonitoringyaml-around-lines-1-to-17-the)
- [x] [In deployments/kubernetes/monitoring.yaml around lines 72-84, the alert](#in-deploymentskubernetesmonitoringyaml-around-lines-72-84-the-alert)
- [x] [In deployments/kubernetes/monitoring.yaml around lines 96-107, the rule](#in-deploymentskubernetesmonitoringyaml-around-lines-96-107-the-rule)
- [x] [In deployments/kubernetes/monitoring.yaml around lines 109 to 119, the alert](#in-deploymentskubernetesmonitoringyaml-around-lines-109-to-119-the-alert)
- [x] [In deployments/kubernetes/monitoring.yaml around line 220, the file is missing a](#in-deploymentskubernetesmonitoringyaml-around-line-220-the-file-is-missing-a)
- [x] [In deployments/README.md around lines 39-44 (and also update occurrences at](#in-deploymentsreadmemd-around-lines-39-44-and-also-update-occurrences-at)
- [x] [In deployments/README.md around lines 80 to 88, the README claims a rate limit](#in-deploymentsreadmemd-around-lines-80-to-88-the-readme-claims-a-rate-limit)
- [x] [In deployments/README.md around lines 93 to 100, the current example uses](#in-deploymentsreadmemd-around-lines-93-to-100-the-current-example-uses)
- [x] [In deployments/README.md around lines 95–100 the secret is created with -n](#in-deploymentsreadmemd-around-lines-95100-the-secret-is-created-with--n)
- [ ] [In deployments/README.md around lines 146-148, the docs mention rate limits but](#in-deploymentsreadmemd-around-lines-146-148-the-docs-mention-rate-limits-but)
- [x] [In docs/00_assessment.md around lines 20–21, the doc currently pins "go-redis](#in-docs00assessmentmd-around-lines-2021-the-doc-currently-pins-go-redis)
- [x] [In docs/07_test_plan.md around lines 27 to 29, replace the vague "Chaos (where](#in-docs07testplanmd-around-lines-27-to-29-replace-the-vague-chaos-where)
- [x] [In docs/07_test_plan.md around lines 41 to 45, the benchmark notes lack](#in-docs07testplanmd-around-lines-41-to-45-the-benchmark-notes-lack)
- [x] [In docs/api/anomaly-radar-slo-budget.md around lines 124 to 176, the duration](#in-docsapianomaly-radar-slo-budgetmd-around-lines-124-to-176-the-duration)
- [x] [In docs/api/anomaly-radar-slo-budget.md around lines 186 to 217, the config](#in-docsapianomaly-radar-slo-budgetmd-around-lines-186-to-217-the-config)
- [x] [In docs/api/canary-deployments.md around lines 15 to 19, the authentication](#in-docsapicanary-deploymentsmd-around-lines-15-to-19-the-authentication)
- [x] [In docs/api/canary-deployments.md around lines 303–345, the JSON mixes units](#in-docsapicanary-deploymentsmd-around-lines-303345-the-json-mixes-units)
- [x] [In docs/api/canary-deployments.md around lines 556 to 592, the Deployment Object](#in-docsapicanary-deploymentsmd-around-lines-556-to-592-the-deployment-object)
- [ ] [In docs/api/capacity-planning-api.md around lines 146 to 158 the example calls](#in-docsapicapacity-planning-apimd-around-lines-146-to-158-the-example-calls)
- [x] [In docs/api/capacity-planning-api.md around lines 311 to 318, the import uses a](#in-docsapicapacity-planning-apimd-around-lines-311-to-318-the-import-uses-a)
- [x] [In docs/api/exactly-once-admin.md around lines 299 to 321, replace the literal](#in-docsapiexactly-once-adminmd-around-lines-299-to-321-replace-the-literal)
- [x] [In docs/PRD.md around lines 162-168 the current metric definition](#in-docsprdmd-around-lines-162-168-the-current-metric-definition)
- [x] [`](#)
- [x] [In AGENTS.md around lines 183 to 193, replace the fake placeholder link](#in-agentsmd-around-lines-183-to-193-replace-the-fake-placeholder-link)
- [x] [In cmd/tui/main.go around line 31, the FlagSet is created with flag.ExitOnError](#in-cmdtuimaingo-around-line-31-the-flagset-is-created-with-flagexitonerror)
- [x] [In cmd/tui/main.go around line 43, the error returned by fs.Parse(os.Args[1:])](#in-cmdtuimaingo-around-line-43-the-error-returned-by-fsparseosargs1)
- [x] [In cmd/tui/main.go around lines 64-66, the code pings Redis and merely logs on](#in-cmdtuimaingo-around-lines-64-66-the-code-pings-redis-and-merely-logs-on)
- [x] [In cmd/tui/main.go around line 68, the TODO leaves many CLI flags un-wired](#in-cmdtuimaingo-around-line-68-the-todo-leaves-many-cli-flags-un-wired)
- [x] [In deployments/admin-api/deploy.sh around lines 1 to 10, the script uses weak](#in-deploymentsadmin-apideploysh-around-lines-1-to-10-the-script-uses-weak)
- [x] [In deployments/admin-api/deploy.sh around lines 31 to 43, the docker tag/push](#in-deploymentsadmin-apideploysh-around-lines-31-to-43-the-docker-tagpush)
- [x] [In deployments/admin-api/deploy.sh around lines 49 to 61, the script creates the](#in-deploymentsadmin-apideploysh-around-lines-49-to-61-the-script-creates-the)
- [x] [In deployments/admin-api/deploy.sh around lines 73-85 (and also apply same](#in-deploymentsadmin-apideploysh-around-lines-73-85-and-also-apply-same)
- [x] [In deployments/admin-api/deploy.sh around lines 170-177, replace the deprecated](#in-deploymentsadmin-apideploysh-around-lines-170-177-replace-the-deprecated)
- [x] [In deployments/admin-api/docker-compose.yaml around line 28, the JWT_SECRET is](#in-deploymentsadmin-apidocker-composeyaml-around-line-28-the-jwtsecret-is)
- [x] [In deployments/admin-api/docker-compose.yaml around line 55 the file is missing](#in-deploymentsadmin-apidocker-composeyaml-around-line-55-the-file-is-missing)
- [x] [In deployments/docker/docker-compose.yaml around lines 36-38 (and also update](#in-deploymentsdockerdocker-composeyaml-around-lines-36-38-and-also-update)
- [x] [In deployments/docker/docker-compose.yaml around lines 121 to 123, the file is](#in-deploymentsdockerdocker-composeyaml-around-lines-121-to-123-the-file-is)
- [x] [In deployments/docker/Dockerfile.admin-api around lines 52-53, the HEALTHCHECK](#in-deploymentsdockerdockerfileadmin-api-around-lines-52-53-the-healthcheck)
- [x] [In deployments/docker/Dockerfile.rbac-token-service around lines 39-40, the COPY](#in-deploymentsdockerdockerfilerbac-token-service-around-lines-39-40-the-copy)
- [x] [In deployments/docker/rbac-configs/resources.yaml around line 231, the file is](#in-deploymentsdockerrbac-configsresourcesyaml-around-line-231-the-file-is)
- [x] [In deployments/docker/rbac-configs/roles.yaml around lines 95 to 102, add a](#in-deploymentsdockerrbac-configsrolesyaml-around-lines-95-to-102-add-a)
- [x] [In deployments/docker/rbac-configs/token-service.yaml around line 28, the](#in-deploymentsdockerrbac-configstoken-serviceyaml-around-line-28-the)
- [x] [In deployments/docker/rbac-configs/token-service.yaml around lines 72-75 the](#in-deploymentsdockerrbac-configstoken-serviceyaml-around-lines-72-75-the)
- [x] [In deployments/docker/rbac-configs/token-service.yaml around line 114, the file](#in-deploymentsdockerrbac-configstoken-serviceyaml-around-line-114-the-file)
- [x] [In deployments/README-RBAC-Deployment.md around lines 19 to 36, replace the](#in-deploymentsreadme-rbac-deploymentmd-around-lines-19-to-36-replace-the)
- [x] [In deployments/README-RBAC-Deployment.md around lines 138-146, the env var table](#in-deploymentsreadme-rbac-deploymentmd-around-lines-138-146-the-env-var-table)
- [x] [In deployments/README-RBAC-Deployment.md around lines 257 to 266, the "STOP at](#in-deploymentsreadme-rbac-deploymentmd-around-lines-257-to-266-the-stop-at)
- [x] [In docs/api/advanced-rate-limiting-api.md around lines 140–155, the](#in-docsapiadvanced-rate-limiting-apimd-around-lines-140155-the)
- [ ] [In docs/api/advanced-rate-limiting-api.md around lines 374 to 381, the note that](#in-docsapiadvanced-rate-limiting-apimd-around-lines-374-to-381-the-note-that)
- [x] [In docs/api/anomaly-radar-slo-budget.md around lines 36 to 38, the Go import](#in-docsapianomaly-radar-slo-budgetmd-around-lines-36-to-38-the-go-import)
- [x] [In docs/api/calendar-view.md around lines 39 to 45, the CalendarView struct uses](#in-docsapicalendar-viewmd-around-lines-39-to-45-the-calendarview-struct-uses)
- [x] [In docs/api/calendar-view.md around lines 193 to 241, the bulk reschedule](#in-docsapicalendar-viewmd-around-lines-193-to-241-the-bulk-reschedule)
- [x] [In docs/api/calendar-view.md around lines 575 to 582, the table exposes numeric](#in-docsapicalendar-viewmd-around-lines-575-to-582-the-table-exposes-numeric)
- [x] [In docs/api/distributed-tracing-integration.md around lines 41 to 50, the sample](#in-docsapidistributed-tracing-integrationmd-around-lines-41-to-50-the-sample)
- [x] [In docs/api/distributed-tracing-integration.md around lines 348–359, the example](#in-docsapidistributed-tracing-integrationmd-around-lines-348359-the-example)
- [ ] [In docs/PRD.md around lines 130-131, the rate-limiter note is inaccurate:](#in-docsprdmd-around-lines-130-131-the-rate-limiter-note-is-inaccurate)
- [ ] [In docs/PRD.md around lines 134-136, the current recommendation of using](#in-docsprdmd-around-lines-134-136-the-current-recommendation-of-using)
- [ ] [In docs/PRD.md around lines 169 to 174, the PRD currently omits a](#in-docsprdmd-around-lines-169-to-174-the-prd-currently-omits-a)
- [x] [docs/SLAPS/worker-reflections/claude-001-reflection.md lines 39-43: the](#docsslapsworker-reflectionsclaude-001-reflectionmd-lines-39-43-the)
- [ ] [In Makefile around lines 35 to 40, there is no clean target; add a PHONY clean](#in-makefile-around-lines-35-to-40-there-is-no-clean-target-add-a-phony-clean)
- [x] [In deployments/admin-api/deploy.sh around lines 123 to 137, the ServiceMonitor](#in-deploymentsadmin-apideploysh-around-lines-123-to-137-the-servicemonitor)
- [x] [In deployments/admin-api/k8s-deployment.yaml around lines 42-43 the jwt-secret](#in-deploymentsadmin-apik8s-deploymentyaml-around-lines-42-43-the-jwt-secret)
- [x] [In deployments/admin-api/k8s-deployment.yaml around line 65, the container image](#in-deploymentsadmin-apik8s-deploymentyaml-around-line-65-the-container-image)
- [x] [In deployments/admin-api/k8s-deployment.yaml around line 197, the file is](#in-deploymentsadmin-apik8s-deploymentyaml-around-line-197-the-file-is)
- [x] [In deployments/docker/docker-compose.yaml around lines 50 to 53, the healthcheck](#in-deploymentsdockerdocker-composeyaml-around-lines-50-to-53-the-healthcheck)
- [x] [In deployments/docker/docker-compose.yaml around lines 86–106 (Prometheus) and](#in-deploymentsdockerdocker-composeyaml-around-lines-86106-prometheus-and)
- [x] [In deployments/docker/rbac-configs/resources.yaml around lines 91 to 205 the](#in-deploymentsdockerrbac-configsresourcesyaml-around-lines-91-to-205-the)
- [x] [deployments/docker/rbac-configs/resources.yaml around lines 205–231: the DLQ](#deploymentsdockerrbac-configsresourcesyaml-around-lines-205231-the-dlq)
- [x] [In deployments/docker/rbac-configs/roles.yaml around lines 19 to 23, the roles](#in-deploymentsdockerrbac-configsrolesyaml-around-lines-19-to-23-the-roles)
- [x] [In deployments/kubernetes/admin-api-deployment.yaml around line 98, the](#in-deploymentskubernetesadmin-api-deploymentyaml-around-line-98-the)
- [x] [In deployments/kubernetes/admin-api-deployment.yaml around lines 190-193 the](#in-deploymentskubernetesadmin-api-deploymentyaml-around-lines-190-193-the)
- [x] [In deployments/kubernetes/admin-api-deployment.yaml around line 271, the file is](#in-deploymentskubernetesadmin-api-deploymentyaml-around-line-271-the-file-is)
- [x] [In deployments/kubernetes/rbac-monitoring.yaml around lines 1 to 18, the](#in-deploymentskubernetesrbac-monitoringyaml-around-lines-1-to-18-the)
- [x] [In deployments/kubernetes/rbac-monitoring.yaml around line 43, the runbook_url](#in-deploymentskubernetesrbac-monitoringyaml-around-line-43-the-runbookurl)
- [x] [In deployments/kubernetes/rbac-monitoring.yaml around lines 354 and 362, the](#in-deploymentskubernetesrbac-monitoringyaml-around-lines-354-and-362-the)
- [x] [In deployments/kubernetes/rbac-monitoring.yaml around line 373, the file is](#in-deploymentskubernetesrbac-monitoringyaml-around-line-373-the-file-is)
- [x] [In deployments/kubernetes/rbac-token-service-deployment.yaml around lines](#in-deploymentskubernetesrbac-token-service-deploymentyaml-around-lines)
- [x] [In deployments/kubernetes/rbac-token-service-deployment.yaml around lines 301 to](#in-deploymentskubernetesrbac-token-service-deploymentyaml-around-lines-301-to)
- [x] [In deployments/kubernetes/rbac-token-service-deployment.yaml around lines 403 to](#in-deploymentskubernetesrbac-token-service-deploymentyaml-around-lines-403-to)
- [x] [In deployments/scripts/deploy-rbac-staging.sh around lines 17-19 (and similarly](#in-deploymentsscriptsdeploy-rbac-stagingsh-around-lines-17-19-and-similarly)
- [x] [In deployments/scripts/deploy-rbac-staging.sh around line 155, the SERVICE_IP](#in-deploymentsscriptsdeploy-rbac-stagingsh-around-line-155-the-serviceip)
- [ ] [In deployments/scripts/deploy-rbac-staging.sh around line 211, the unquoted](#in-deploymentsscriptsdeploy-rbac-stagingsh-around-line-211-the-unquoted)
- [x] [In deployments/scripts/deploy-staging.sh around line 71, the docker build](#in-deploymentsscriptsdeploy-stagingsh-around-line-71-the-docker-build)
- [x] [In deployments/scripts/deploy-staging.sh around lines 73, 85, 99, and 122, the](#in-deploymentsscriptsdeploy-stagingsh-around-lines-73-85-99-and-122-the)
- [x] [In deployments/scripts/deploy-staging.sh around line 83 (and also at lines 112](#in-deploymentsscriptsdeploy-stagingsh-around-line-83-and-also-at-lines-112)
- [x] [In deployments/scripts/deploy-staging.sh around line 225, the conditional uses](#in-deploymentsscriptsdeploy-stagingsh-around-line-225-the-conditional-uses)
- [x] [In deployments/scripts/health-check-rbac.sh around line 238, the current command](#in-deploymentsscriptshealth-check-rbacsh-around-line-238-the-current-command)
- [x] [In deployments/scripts/setup-monitoring.sh around lines 17 to 31 the logging](#in-deploymentsscriptssetup-monitoringsh-around-lines-17-to-31-the-logging)
- [x] [In deployments/scripts/setup-monitoring.sh around lines 132 to 180, the](#in-deploymentsscriptssetup-monitoringsh-around-lines-132-to-180-the)
- [x] [In deployments/scripts/test-staging-deployment.sh around lines 63-68 (and](#in-deploymentsscriptstest-staging-deploymentsh-around-lines-63-68-and)
- [x] [In deployments/scripts/test-staging-deployment.sh around lines 331-333, the](#in-deploymentsscriptstest-staging-deploymentsh-around-lines-331-333-the)
- [x] [In deployments/scripts/test-staging-deployment.sh around lines 451-453 (and](#in-deploymentsscriptstest-staging-deploymentsh-around-lines-451-453-and)
- [x] [In docs/FEATURE_ENHANCEMENT_AGENT_PROMPT.md around lines 32 to 35, the](#in-docsfeatureenhancementagentpromptmd-around-lines-32-to-35-the)
- [x] [In docs/FEATURE_ENHANCEMENT_AGENT_PROMPT.md around lines 156–180 the color](#in-docsfeatureenhancementagentpromptmd-around-lines-156180-the-color)
- [ ] [In docs/TUI/README.md around lines 15 to 21, the README shows verified images](#in-docstuireadmemd-around-lines-15-to-21-the-readme-shows-verified-images)
- [x] [In Makefile around line 13, remove the unnecessary GO111MODULE=on prefix from](#in-makefile-around-line-13-remove-the-unnecessary-go111moduleon-prefix-from)
- [x] [In Makefile around lines 16 to 18, the TUI build target uses a different go](#in-makefile-around-lines-16-to-18-the-tui-build-target-uses-a-different-go)
- [x] [In .githooks/pre-commit around lines 7 to 15, the hook assumes python3 exists](#in-githookspre-commit-around-lines-7-to-15-the-hook-assumes-python3-exists)
- [x] [In .github/workflows/update-progress.yml lines 1-6, YAMLLint flagged the](#in-githubworkflowsupdate-progressyml-lines-1-6-yamllint-flagged-the)
- [x] [In .github/workflows/update-progress.yml around lines 8 to 10, the workflow](#in-githubworkflowsupdate-progressyml-around-lines-8-to-10-the-workflow)
- [x] [.github/workflows/update-progress.yml lines 31-41: the workflow uses unguarded](#githubworkflowsupdate-progressyml-lines-31-41-the-workflow-uses-unguarded)
- [x] [.github/workflows/update-progress.yml around lines 43-47 contains an extra](#githubworkflowsupdate-progressyml-around-lines-43-47-contains-an-extra)
- [x] [In cmd/job-queue-system/main.go around lines 85 to 92, the code always starts](#in-cmdjob-queue-systemmaingo-around-lines-85-to-92-the-code-always-starts)
- [x] [In cmd/job-queue-system/main.go around lines 112–114, the background metrics](#in-cmdjob-queue-systemmaingo-around-lines-112114-the-background-metrics)
- [x] [In cmd/job-queue-system/main.go around lines 142 to 149, the admin handling is](#in-cmdjob-queue-systemmaingo-around-lines-142-to-149-the-admin-handling-is)
- [x] [In cmd/job-queue-system/main.go around lines 187-188, the purge-all branch](#in-cmdjob-queue-systemmaingo-around-lines-187-188-the-purge-all-branch)
- [x] [In deployments/kubernetes/rbac-monitoring.yaml around line 35 (and also review](#in-deploymentskubernetesrbac-monitoringyaml-around-line-35-and-also-review)
- [x] [In Dockerfile around line 3, the build stage uses golang:1.23 while CI/docs](#in-dockerfile-around-line-3-the-build-stage-uses-golang123-while-cidocs)
- [x] [In docs/15_promotion_checklists.md around lines 21–33, the promotion checklist](#in-docs15promotionchecklistsmd-around-lines-2133-the-promotion-checklist)
- [x] [In docs/api/admin-api.md around lines 41 to 44, the configuration shows a single](#in-docsapiadmin-apimd-around-lines-41-to-44-the-configuration-shows-a-single)
- [x] [In docs/api/admin-api.md around lines 106 to 132, the queue parameter](#in-docsapiadmin-apimd-around-lines-106-to-132-the-queue-parameter)
- [x] [In docs/api/admin-api.md around lines 260 to 268, the CORS guidance currently](#in-docsapiadmin-apimd-around-lines-260-to-268-the-cors-guidance-currently)
- [x] [docs/api/event-hooks.md around lines 110 to 124: the HMAC signature currently](#docsapievent-hooksmd-around-lines-110-to-124-the-hmac-signature-currently)
- [x] [In docs/api/event-hooks.md around lines 246 to 264, the docs currently describe](#in-docsapievent-hooksmd-around-lines-246-to-264-the-docs-currently-describe)
- [x] [In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 41-44, the docs show an](#in-eventhookstestdocumentationmd-around-lines-41-44-the-docs-show-an)
- [x] [In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 61 to 65 the example test](#in-eventhookstestdocumentationmd-around-lines-61-to-65-the-example-test)
- [x] [In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 86-90, the docs currently](#in-eventhookstestdocumentationmd-around-lines-86-90-the-docs-currently)
- [x] [In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 112 to 116, the example](#in-eventhookstestdocumentationmd-around-lines-112-to-116-the-example)
- [x] [In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 136 to 139, the example test](#in-eventhookstestdocumentationmd-around-lines-136-to-139-the-example-test)
- [x] [`](#)
- [x] [In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 216-224, the docs incorrectly](#in-eventhookstestdocumentationmd-around-lines-216-224-the-docs-incorrectly)
- [x] [In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 228-235, the docs use a](#in-eventhookstestdocumentationmd-around-lines-228-235-the-docs-use-a)
- [x] [In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 239 to 244, the benchmark](#in-eventhookstestdocumentationmd-around-lines-239-to-244-the-benchmark)
- [x] [In create_postmortem_tasks.py around lines 107 to 108, the dependencies array is](#in-createpostmortemtaskspy-around-lines-107-to-108-the-dependencies-array-is)
- [x] [In create_postmortem_tasks.py around lines 114 to 117, the two os.makedirs calls](#in-createpostmortemtaskspy-around-lines-114-to-117-the-two-osmakedirs-calls)
- [x] [In demos/lipgloss-transformation.tape lines 1–278, the demo invokes a](#in-demoslipgloss-transformationtape-lines-1278-the-demo-invokes-a)
- [x] [In deploy/deploy/data/test.txt lines 1-1 the test data is incorrectly placed](#in-deploydeploydatatesttxt-lines-1-1-the-test-data-is-incorrectly-placed)
- [x] [In deploy/deploy/data/test.txt around lines 1 to 1, the file contains only the](#in-deploydeploydatatesttxt-around-lines-1-to-1-the-file-contains-only-the)
- [x] [In deploy/grafana/dashboards/work-queue.json lines 1-37, the dashboard currently](#in-deploygrafanadashboardswork-queuejson-lines-1-37-the-dashboard-currently)
- [x] [In deployments/admin-api/Dockerfile around lines 16 to 18, the go build](#in-deploymentsadmin-apidockerfile-around-lines-16-to-18-the-go-build)
- [x] [In deployments/admin-api/Dockerfile around lines 20 to 23, the image currently](#in-deploymentsadmin-apidockerfile-around-lines-20-to-23-the-image-currently)
- [x] [In deployments/admin-api/Dockerfile around lines 26 to 31, the Dockerfile](#in-deploymentsadmin-apidockerfile-around-lines-26-to-31-the-dockerfile)
- [x] [In deployments/docker/Dockerfile.admin-api around lines 20-21, the Go build](#in-deploymentsdockerdockerfileadmin-api-around-lines-20-21-the-go-build)
- [x] [In deployments/docker/Dockerfile.admin-api around lines 38-40, the COPY line](#in-deploymentsdockerdockerfileadmin-api-around-lines-38-40-the-copy-line)
- [x] [In docs/03_milestones.md around lines 11-13 (and similarly at lines 46-51), the](#in-docs03milestonesmd-around-lines-11-13-and-similarly-at-lines-46-51-the)
- [ ] [In docs/api/anomaly-radar-slo-budget.md around lines 1–15, add a "Versioning &](#in-docsapianomaly-radar-slo-budgetmd-around-lines-115-add-a-versioning-)
- [x] [In docs/api/anomaly-radar-slo-budget.md around lines 36-38 (and also apply the](#in-docsapianomaly-radar-slo-budgetmd-around-lines-36-38-and-also-apply-the)
- [x] [In docs/api/anomaly-radar-slo-budget.md around lines 36 to 38, the docs instruct](#in-docsapianomaly-radar-slo-budgetmd-around-lines-36-to-38-the-docs-instruct)
- [ ] [In docs/api/anomaly-radar-slo-budget.md around lines 51 to 58, the docs present](#in-docsapianomaly-radar-slo-budgetmd-around-lines-51-to-58-the-docs-present)
- [ ] [In docs/api/anomaly-radar-slo-budget.md around lines 81 to 89, the SLOConfig](#in-docsapianomaly-radar-slo-budgetmd-around-lines-81-to-89-the-sloconfig)
- [ ] [In docs/api/anomaly-radar-slo-budget.md around lines 94-102 (and also apply same](#in-docsapianomaly-radar-slo-budgetmd-around-lines-94-102-and-also-apply-same)
- [ ] [In docs/api/anomaly-radar-slo-budget.md around line 124, there is no](#in-docsapianomaly-radar-slo-budgetmd-around-line-124-there-is-no)
- [x] [In docs/api/anomaly-radar-slo-budget.md around lines 124-176 (and apply same](#in-docsapianomaly-radar-slo-budgetmd-around-lines-124-176-and-apply-same)
- [x] [docs/api/anomaly-radar-slo-budget.md around lines 128 to 176: the response](#docsapianomaly-radar-slo-budgetmd-around-lines-128-to-176-the-response)
- [ ] [In docs/api/anomaly-radar-slo-budget.md around lines 132-135 (and also apply](#in-docsapianomaly-radar-slo-budgetmd-around-lines-132-135-and-also-apply)
- [x] [In docs/api/anomaly-radar-slo-budget.md around lines 220 to 238, the PUT](#in-docsapianomaly-radar-slo-budgetmd-around-lines-220-to-238-the-put)
- [x] [In docs/api/anomaly-radar-slo-budget.md around lines 246-249, the query param](#in-docsapianomaly-radar-slo-budgetmd-around-lines-246-249-the-query-param)
- [x] [In docs/api/anomaly-radar-slo-budget.md around lines 250-266 (and similarly for](#in-docsapianomaly-radar-slo-budgetmd-around-lines-250-266-and-similarly-for)
- [ ] [In docs/api/anomaly-radar-slo-budget.md around lines 299 to 327, the response](#in-docsapianomaly-radar-slo-budgetmd-around-lines-299-to-327-the-response)
- [x] [In docs/api/anomaly-radar-slo-budget.md around lines 371 to 374, the health](#in-docsapianomaly-radar-slo-budgetmd-around-lines-371-to-374-the-health)
- [x] [In docs/api/anomaly-radar-slo-budget.md around lines 375 to 383, the Start/Stop](#in-docsapianomaly-radar-slo-budgetmd-around-lines-375-to-383-the-startstop)
- [x] [In docs/api/anomaly-radar-slo-budget.md around lines 435 to 441, the "Batch](#in-docsapianomaly-radar-slo-budgetmd-around-lines-435-to-441-the-batch)
- [x] [In docs/api/anomaly-radar-slo-budget.md around lines 488 to 503, the Prometheus](#in-docsapianomaly-radar-slo-budgetmd-around-lines-488-to-503-the-prometheus)
- [x] [In docs/api/anomaly-radar-slo-budget.md around lines 519 to 521, the README](#in-docsapianomaly-radar-slo-budgetmd-around-lines-519-to-521-the-readme)
- [x] [In docs/api/chaos-harness.md around lines 259 to 283, the example imports an](#in-docsapichaos-harnessmd-around-lines-259-to-283-the-example-imports-an)
- [x] [In docs/api/dlq-remediation-ui.md around lines 7 to 10, the documentation](#in-docsapidlq-remediation-uimd-around-lines-7-to-10-the-documentation)
- [x] [In docs/api/dlq-remediation-ui.md around lines 25 to 39, the docs currently](#in-docsapidlq-remediation-uimd-around-lines-25-to-39-the-docs-currently)
- [x] [In docs/api/dlq-remediation-ui.md around lines 223-241, the purge-all endpoint](#in-docsapidlq-remediation-uimd-around-lines-223-241-the-purge-all-endpoint)
- [x] [In docs/api/dlq-remediation-ui.md around lines 247 to 257, the JSON example](#in-docsapidlq-remediation-uimd-around-lines-247-to-257-the-json-example)
- [x] [In docs/api/dlq-remediation-ui.md around lines 299 to 307, the HTTP status codes](#in-docsapidlq-remediation-uimd-around-lines-299-to-307-the-http-status-codes)
- [x] [In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 246 to 262, the listed](#in-eventhookstestdocumentationmd-around-lines-246-to-262-the-listed)
- [x] [In .github/workflows/markdownlint.yml around line 3 the reserved YAML key on is](#in-githubworkflowsmarkdownlintyml-around-line-3-the-reserved-yaml-key-on-is)
- [x] [.github/workflows/update-progress.yml lines 16-24: the workflow currently](#githubworkflowsupdate-progressyml-lines-16-24-the-workflow-currently)
- [ ] [In .gitignore around lines 28-33, you need to prevent accidental commits of](#in-gitignore-around-lines-28-33-you-need-to-prevent-accidental-commits-of)
- [x] [In append_metadata.py around lines 30 to 33, format_list currently returns "-](#in-appendmetadatapy-around-lines-30-to-33-formatlist-currently-returns--)
- [x] [In append_metadata.py around lines 35 to 59 (and also apply same pattern to](#in-appendmetadatapy-around-lines-35-to-59-and-also-apply-same-pattern-to)
- [x] [In BUGS.md around lines 12 to 25, the heartbeat snippet is unsafe: it ignores](#in-bugsmd-around-lines-12-to-25-the-heartbeat-snippet-is-unsafe-it-ignores)
- [x] [In BUGS.md around lines 32-38, the current code scans Redis with pattern](#in-bugsmd-around-lines-32-38-the-current-code-scans-redis-with-pattern)
- [x] [In BUGS.md around lines 45–53 the scheduler mover uses ZRANGEBYSCORE + ZREM +](#in-bugsmd-around-lines-4553-the-scheduler-mover-uses-zrangebyscore--zrem-)
- [x] [In BUGS.md around lines 55–61, the advice to write NDJSON ledger files to local](#in-bugsmd-around-lines-5561-the-advice-to-write-ndjson-ledger-files-to-local)
- [x] [In claude_worker.py around lines 29 to 34, only self.my_dir is created but other](#in-claudeworkerpy-around-lines-29-to-34-only-selfmydir-is-created-but-other)
- [x] [In cmd/job-queue-system/main.go around lines 159-161 (and similarly at 169-171](#in-cmdjob-queue-systemmaingo-around-lines-159-161-and-similarly-at-169-171)
- [x] [In create_postmortem_tasks.py around lines 1-3, the code appends "Z" to a naive](#in-createpostmortemtaskspy-around-lines-1-3-the-code-appends-z-to-a-naive)
- [x] [In create_postmortem_tasks.py around lines 15 to 16, the code appends "Z" to](#in-createpostmortemtaskspy-around-lines-15-to-16-the-code-appends-z-to)
- [x] [In create_postmortem_tasks.py around lines 27 to 39, the criteria strings still](#in-createpostmortemtaskspy-around-lines-27-to-39-the-criteria-strings-still)
- [x] [In create_postmortem_tasks.py around lines 69-70, the timestamp is created with](#in-createpostmortemtaskspy-around-lines-69-70-the-timestamp-is-created-with)
- [x] [In create_review_tasks.py around lines 10-11, the check for "duplicate" is](#in-createreviewtaskspy-around-lines-10-11-the-check-for-duplicate-is)
- [x] [In create_review_tasks.py around lines 14 to 21, the code uses a bare except](#in-createreviewtaskspy-around-lines-14-to-21-the-code-uses-a-bare-except)
- [x] [In demos/responsive-tui.tape around lines 72-73 (and also at 129-130, 214-215,](#in-demosresponsive-tuitape-around-lines-72-73-and-also-at-129-130-214-215)
- [x] [In dependency_analysis.py around lines 23–166, there’s a naming inconsistency:](#in-dependencyanalysispy-around-lines-23166-theres-a-naming-inconsistency)
- [x] [In deployments/admin-api/k8s-redis.yaml around lines 16 to 52, the Pod and](#in-deploymentsadmin-apik8s-redisyaml-around-lines-16-to-52-the-pod-and)
- [x] [In deployments/docker/Dockerfile.rbac-token-service around lines 25-27 (and also](#in-deploymentsdockerdockerfilerbac-token-service-around-lines-25-27-and-also)
- [x] [In docs/00_assessment.md around line 3, the "Last updated: 2025-09-12" header is](#in-docs00assessmentmd-around-line-3-the-last-updated-2025-09-12-header-is)
- [x] [In docs/02_release_plan.md around lines 6–7, the release plan text needs](#in-docs02releaseplanmd-around-lines-67-the-release-plan-text-needs)
- [x] [In docs/10_risk_register.md around line 3, the "Last updated" timestamp is stale](#in-docs10riskregistermd-around-line-3-the-last-updated-timestamp-is-stale)
- [x] [In docs/api/admin-api.md around lines 356 to 359, the docs mention a “minimum](#in-docsapiadmin-apimd-around-lines-356-to-359-the-docs-mention-a-minimum)
- [ ] [In docs/api/anomaly-radar-slo-budget.md around lines 76 to 80, the repo has](#in-docsapianomaly-radar-slo-budgetmd-around-lines-76-to-80-the-repo-has)
- [x] [In docs/api/canary-deployments.md around lines 7 to 11 (and also apply the same](#in-docsapicanary-deploymentsmd-around-lines-7-to-11-and-also-apply-the-same)
- [ ] [In docs/api/dlq-remediation-pipeline.md around lines 761 to 858, the notify](#in-docsapidlq-remediation-pipelinemd-around-lines-761-to-858-the-notify)
- [x] [In docs/api/exactly-once-admin.md around lines 25-33 (and also apply the same](#in-docsapiexactly-once-adminmd-around-lines-25-33-and-also-apply-the-same)
- [x] [In docs/SLAPS/coordinator-observations.md around lines 114-116 (and also apply](#in-docsslapscoordinator-observationsmd-around-lines-114-116-and-also-apply)
- [x] [In docs/SLAPS/coordinator-observations.md around lines 121 to 130, the](#in-docsslapscoordinator-observationsmd-around-lines-121-to-130-the)
- [x] [docs/SLAPS/coordinator-observations.md around lines 249-251: the document](#docsslapscoordinator-observationsmd-around-lines-249-251-the-document)
- [x] [In docs/YOU ARE WORKER 6.rb around lines 1 to 5, the file contains non-Ruby](#in-docsyou-are-worker-6rb-around-lines-1-to-5-the-file-contains-non-ruby)
- [x] [In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 314 to 320, the examples use](#in-eventhookstestdocumentationmd-around-lines-314-to-320-the-examples-use)
- [x] [In AGENTS.md around lines 10 to 41, the table of contents uses Obsidian-style](#in-agentsmd-around-lines-10-to-41-the-table-of-contents-uses-obsidian-style)
- [x] [In deployments/admin-api/k8s-deployment.yaml around lines 62 to 116, the](#in-deploymentsadmin-apik8s-deploymentyaml-around-lines-62-to-116-the)
- [x] [In deployments/admin-api/k8s-deployment.yaml around lines 117 to 159, the](#in-deploymentsadmin-apik8s-deploymentyaml-around-lines-117-to-159-the)
- [x] [In deployments/docker/docker-compose.yaml around lines 105 to 109 the compose](#in-deploymentsdockerdocker-composeyaml-around-lines-105-to-109-the-compose)
- [x] [In deployments/docker/rbac-configs/resources.yaml around lines 91-104 (and also](#in-deploymentsdockerrbac-configsresourcesyaml-around-lines-91-104-and-also)
- [x] [In deployments/docker/rbac-configs/token-service.yaml around lines 21 to 24, the](#in-deploymentsdockerrbac-configstoken-serviceyaml-around-lines-21-to-24-the)
- [x] [In deployments/kubernetes/admin-api-deployment.yaml around lines 90 to 100, the](#in-deploymentskubernetesadmin-api-deploymentyaml-around-lines-90-to-100-the)
- [x] [In deployments/kubernetes/monitoring.yaml around lines 35 to 38, the alert](#in-deploymentskubernetesmonitoringyaml-around-lines-35-to-38-the-alert)
- [x] [In deployments/kubernetes/rbac-monitoring.yaml around lines 45–54, the expr](#in-deploymentskubernetesrbac-monitoringyaml-around-lines-4554-the-expr)
- [x] [In deployments/kubernetes/rbac-monitoring.yaml around lines 56 to 75, the](#in-deploymentskubernetesrbac-monitoringyaml-around-lines-56-to-75-the)
- [x] [In deployments/kubernetes/rbac-token-service-deployment.yaml around lines](#in-deploymentskubernetesrbac-token-service-deploymentyaml-around-lines)
- [x] [In deployments/kubernetes/rbac-token-service-deployment.yaml around lines](#in-deploymentskubernetesrbac-token-service-deploymentyaml-around-lines)
- [x] [In deployments/scripts/deploy-staging.sh around lines 161-166 (and also apply](#in-deploymentsscriptsdeploy-stagingsh-around-lines-161-166-and-also-apply)
- [x] [In deployments/scripts/health-check-rbac.sh around lines 173 to 191, replace the](#in-deploymentsscriptshealth-check-rbacsh-around-lines-173-to-191-replace-the)
- [ ] [In deployments/scripts/setup-monitoring.sh around lines 225-226, the kill call](#in-deploymentsscriptssetup-monitoringsh-around-lines-225-226-the-kill-call)
- [ ] [In docs/api/dlq-remediation-pipeline.md around lines 149-171, the example](#in-docsapidlq-remediation-pipelinemd-around-lines-149-171-the-example)
- [x] [In docs/api/dlq-remediation-pipeline.md around lines 676 to 687, the documented](#in-docsapidlq-remediation-pipelinemd-around-lines-676-to-687-the-documented)
- [x] [In docs/api/dlq-remediation-pipeline.md around lines 701 to 712, the](#in-docsapidlq-remediation-pipelinemd-around-lines-701-to-712-the)
- [x] [In docs/api/dlq-remediation-pipeline.md around lines 713 to 731, the docs](#in-docsapidlq-remediation-pipelinemd-around-lines-713-to-731-the-docs)
- [x] [In docs/api/dlq-remediation-ui.md around lines 432 to 436, the purge-all example](#in-docsapidlq-remediation-uimd-around-lines-432-to-436-the-purge-all-example)
- [x] [docs/SLAPS/worker-reflections/claude-008-reflection.md lines 1-16: add a YAML](#docsslapsworker-reflectionsclaude-008-reflectionmd-lines-1-16-add-a-yaml)
- [ ] [In README.md lines 44-48 the build section targets Go 1.25, but go.mod (line 3)](#in-readmemd-lines-44-48-the-build-section-targets-go-125-but-gomod-line-3)
- [ ] [README.md around lines 70 to 86: new users will hit missing Go modules when](#readmemd-around-lines-70-to-86-new-users-will-hit-missing-go-modules-when)
- [x] [.github/workflows/changelog.yml lines 20-29: workflow currently references](#githubworkflowschangelogyml-lines-20-29-workflow-currently-references)
- [ ] [In .github/workflows/changelog.yml around lines 20 to 24, the checkout step](#in-githubworkflowschangelogyml-around-lines-20-to-24-the-checkout-step)
- [ ] [.github/workflows/changelog.yml around lines 35 to 42: the workflow currently](#githubworkflowschangelogyml-around-lines-35-to-42-the-workflow-currently)
- [ ] [.github/workflows/ci.yml around line 27: CI is using go-version '1.25.x' while](#githubworkflowsciyml-around-line-27-ci-is-using-go-version-125x-while)
- [ ] [.github/workflows/ci.yml around lines 54 to 62: the CI job uses Bash-specific](#githubworkflowsciyml-around-lines-54-to-62-the-ci-job-uses-bash-specific)
- [ ] [In .github/workflows/goreleaser.yml around lines 13–16, the release job is](#in-githubworkflowsgoreleaseryml-around-lines-1316-the-release-job-is)
- [ ] [.github/workflows/goreleaser.yml around lines 39 to 45: the workflow uses](#githubworkflowsgoreleaseryml-around-lines-39-to-45-the-workflow-uses)
- [x] [.github/workflows/markdownlint.yml lines 4-6: the workflow is currently](#githubworkflowsmarkdownlintyml-lines-4-6-the-workflow-is-currently)
- [ ] [.github/workflows/markdownlint.yml lines 15-18: the workflow lacks a job timeout](#githubworkflowsmarkdownlintyml-lines-15-18-the-workflow-lacks-a-job-timeout)
- [ ] [.github/workflows/markdownlint.yml around line 17: the workflow uses the](#githubworkflowsmarkdownlintyml-around-line-17-the-workflow-uses-the)
- [x] [.github/workflows/markdownlint.yml lines 20-21: the checkout action uses](#githubworkflowsmarkdownlintyml-lines-20-21-the-checkout-action-uses)
- [ ] [In .goreleaser.yaml around lines 8 to 13, the build configuration lacks](#in-goreleaseryaml-around-lines-8-to-13-the-build-configuration-lacks)
- [ ] [In .markdownlint.yaml around lines 10 to 13, the MD026 punctuation list includes](#in-markdownlintyaml-around-lines-10-to-13-the-md026-punctuation-list-includes)
- [ ] [.vscode/extensions.json around lines 2 to 7: the workspace recommendations list](#vscodeextensionsjson-around-lines-2-to-7-the-workspace-recommendations-list)
- [ ] [.vscode/extensions.json around line 7: the unwantedRecommendations array is](#vscodeextensionsjson-around-line-7-the-unwantedrecommendations-array-is)
- [ ] [In .vscode/settings.json around lines 1-8, YAML files aren't configured to](#in-vscodesettingsjson-around-lines-1-8-yaml-files-arent-configured-to)
- [ ] [.vscode/settings.json lines 1-8: trim any trailing whitespace on each line and](#vscodesettingsjson-lines-1-8-trim-any-trailing-whitespace-on-each-line-and)
- [ ] [In .vscode/settings.json around line 19, the go.testFlags array currently uses](#in-vscodesettingsjson-around-line-19-the-gotestflags-array-currently-uses)
- [ ] [In CHANGELOG.md around lines 9 to 16, replace all placeholder PR tags "[#PR?]"](#in-changelogmd-around-lines-9-to-16-replace-all-placeholder-pr-tags-pr)
- [ ] [In CHANGELOG.md around line 20, the entry "Smarter rate limiting that sleeps](#in-changelogmd-around-line-20-the-entry-smarter-rate-limiting-that-sleeps)
- [ ] [In CHANGELOG.md around line 38, remove the pseudo‑directive line](#in-changelogmd-around-line-38-remove-the-pseudodirective-line)
- [ ] [In create_review_tasks.py around lines 57 and 91, the coverage threshold is](#in-createreviewtaskspy-around-lines-57-and-91-the-coverage-threshold-is)
- [ ] [In demos/responsive-tui.tape around lines 4-6, the script sets Theme "Tokyo](#in-demosresponsive-tuitape-around-lines-4-6-the-script-sets-theme-tokyo)
- [ ] [In demos/responsive-tui.tape around line 9, the TypingSpeed is set to 80ms which](#in-demosresponsive-tuitape-around-line-9-the-typingspeed-is-set-to-80ms-which)
- [ ] [In demos/responsive-tui.tape around line 10, stop hard-coding zsh; change the](#in-demosresponsive-tuitape-around-line-10-stop-hard-coding-zsh-change-the)
- [ ] [In demos/responsive-tui.tape around lines 12 to 16, the demo relies on the](#in-demosresponsive-tuitape-around-lines-12-to-16-the-demo-relies-on-the)
- [ ] [In demos/responsive-tui.tape around lines 28-29 (also apply the same fix at](#in-demosresponsive-tuitape-around-lines-28-29-also-apply-the-same-fix-at)
- [ ] [In demos/responsive-tui.tape around lines 30 to 73, the script emits an](#in-demosresponsive-tuitape-around-lines-30-to-73-the-script-emits-an)
- [ ] [In dependency_analysis.py around lines 213-218, the entry lists "json_editor" as](#in-dependencyanalysispy-around-lines-213-218-the-entry-lists-jsoneditor-as)
- [ ] [In auto_commit.sh around lines 5 to 16, there is no preflight check for git](#in-autocommitsh-around-lines-5-to-16-there-is-no-preflight-check-for-git)
- [ ] [In auto_commit.sh around lines 51 to 61, the commit message is built via a](#in-autocommitsh-around-lines-51-to-61-the-commit-message-is-built-via-a)
- [ ] [In auto_commit.sh around lines 62 to 75, the script always pushes to origin and](#in-autocommitsh-around-lines-62-to-75-the-script-always-pushes-to-origin-and)
- [ ] [In deploy/data/test.txt (lines 1-1) this stray test artifact should not be](#in-deploydatatesttxt-lines-1-1-this-stray-test-artifact-should-not-be)
- [ ] [In deploy/data/test.txt lines 1-1, this test file must never be included in](#in-deploydatatesttxt-lines-1-1-this-test-file-must-never-be-included-in)
- [ ] [In deploy/grafana/dashboards/work-queue.json around lines 10 to 12, the PromQL](#in-deploygrafanadashboardswork-queuejson-around-lines-10-to-12-the-promql)
- [ ] [deploy/grafana/dashboards/work-queue.json around lines 65-86: the Stat panel](#deploygrafanadashboardswork-queuejson-around-lines-65-86-the-stat-panel)
- [ ] [In deployments/admin-api/k8s-redis.yaml around lines 7 to 12, the ServiceAccount](#in-deploymentsadmin-apik8s-redisyaml-around-lines-7-to-12-the-serviceaccount)
- [ ] [In deployments/admin-api/k8s-redis.yaml around lines 31-34 (and also lines](#in-deploymentsadmin-apik8s-redisyaml-around-lines-31-34-and-also-lines)
- [ ] [In deployments/admin-api/k8s-redis.yaml around lines 40 to 45, the Redis command](#in-deploymentsadmin-apik8s-redisyaml-around-lines-40-to-45-the-redis-command)
- [ ] [In deployments/admin-api/k8s-redis.yaml around lines 66 to 79, the](#in-deploymentsadmin-apik8s-redisyaml-around-lines-66-to-79-the)
- [ ] [In docs/01_product_roadmap.md around lines 54 to 56, the dependency references](#in-docs01productroadmapmd-around-lines-54-to-56-the-dependency-references)
- [ ] [In docs/13_release_versioning.md around lines 17–19, the changelog rules are](#in-docs13releaseversioningmd-around-lines-1719-the-changelog-rules-are)
- [ ] [In docs/13_release_versioning.md around line 23, the phrase "Ensure CI green;](#in-docs13releaseversioningmd-around-line-23-the-phrase-ensure-ci-green)
- [ ] [In docs/13_release_versioning.md around line 26, the doc currently says to](#in-docs13releaseversioningmd-around-line-26-the-doc-currently-says-to)
- [ ] [In docs/13_release_versioning.md around lines 27 to 32, the instructions](#in-docs13releaseversioningmd-around-lines-27-to-32-the-instructions)
- [ ] [In docs/SLAPS/worker-reflections/claude-006-reflection.md around lines 8 to 12,](#in-docsslapsworker-reflectionsclaude-006-reflectionmd-around-lines-8-to-12)
- [ ] [In docs/SLAPS/worker-reflections/claude-006-reflection.md around line 12, the](#in-docsslapsworker-reflectionsclaude-006-reflectionmd-around-line-12-the)
- [ ] [In docs/SLAPS/worker-reflections/claude-006-reflection.md around lines 20-23](#in-docsslapsworker-reflectionsclaude-006-reflectionmd-around-lines-20-23)
- [ ] [docs/SLAPS/worker-reflections/claude-006-reflection.md around line 117: there is](#docsslapsworker-reflectionsclaude-006-reflectionmd-around-line-117-there-is)
- [x] [In docs/YOU ARE WORKER 6.rb around lines 3 to 4, the sentence "You are a worker](#in-docsyou-are-worker-6rb-around-lines-3-to-4-the-sentence-you-are-a-worker)
- [ ] [In docs/YOU ARE WORKER 6.rb around line 28, the mv command may mis-handle](#in-docsyou-are-worker-6rb-around-line-28-the-mv-command-may-mis-handle)
- [ ] [In append_metadata.py around lines 57 to 113, remove the hard-coded](#in-appendmetadatapy-around-lines-57-to-113-remove-the-hard-coded)
- [ ] [In claude_worker.py lines 1-5, the module currently has a short descriptive](#in-claudeworkerpy-lines-1-5-the-module-currently-has-a-short-descriptive)
- [ ] [In claude_worker.py around lines 7 to 16, replace ad-hoc prints with a proper](#in-claudeworkerpy-around-lines-7-to-16-replace-ad-hoc-prints-with-a-proper)
- [ ] [In claude_worker.py around lines 121-123, the unknown-status branch currently](#in-claudeworkerpy-around-lines-121-123-the-unknown-status-branch-currently)
- [ ] [In claude_worker.py around lines 124-127, the except block for](#in-claudeworkerpy-around-lines-124-127-the-except-block-for)
- [ ] [In claude_worker.py around lines 163 to 167 the _persist_task function writes](#in-claudeworkerpy-around-lines-163-to-167-the-persisttask-function-writes)
- [ ] [In claude_worker.py around lines 171 to 180, the argument range check is done](#in-claudeworkerpy-around-lines-171-to-180-the-argument-range-check-is-done)
- [ ] [In config/config.example.yaml around lines 54 to 67, the config references](#in-configconfigexampleyaml-around-lines-54-to-67-the-config-references)
- [ ] [In config/config.example.yaml around lines 72 to 75, the example DSN includes a](#in-configconfigexampleyaml-around-lines-72-to-75-the-example-dsn-includes-a)
- [ ] [In create_review_tasks.py around line 6 (and also update annotations at 53-54](#in-createreviewtaskspy-around-line-6-and-also-update-annotations-at-53-54)
- [ ] [In create_review_tasks.py around lines 35 to 46, fromisoformat() currently fails](#in-createreviewtaskspy-around-lines-35-to-46-fromisoformat-currently-fails)
- [ ] [In create_review_tasks.py around lines 160 to 162, the file is opened without an](#in-createreviewtaskspy-around-lines-160-to-162-the-file-is-opened-without-an)
- [ ] [In demos/lipgloss-transformation.tape around lines 22 to 43, the test currently](#in-demoslipgloss-transformationtape-around-lines-22-to-43-the-test-currently)
- [ ] [In dependency_analysis.py around lines 326 to 338, validate_dependencies is](#in-dependencyanalysispy-around-lines-326-to-338-validatedependencies-is)
- [x] [In deployments/admin-api/monitoring.yaml around lines 1-66, this file duplicates](#in-deploymentsadmin-apimonitoringyaml-around-lines-1-66-this-file-duplicates)
- [ ] [In deployments/admin-api/monitoring.yaml around lines 11 to 18, the alert](#in-deploymentsadmin-apimonitoringyaml-around-lines-11-to-18-the-alert)
- [ ] [In deployments/admin-api/monitoring.yaml around lines 29 to 36, the PromQL uses](#in-deploymentsadmin-apimonitoringyaml-around-lines-29-to-36-the-promql-uses)
- [ ] [In deployments/admin-api/monitoring.yaml around lines 73-76 (and also lines](#in-deploymentsadmin-apimonitoringyaml-around-lines-73-76-and-also-lines)
- [x] [In docs/09_requirements.md around lines 47 to 49, the metrics list needs](#in-docs09requirementsmd-around-lines-47-to-49-the-metrics-list-needs)
- [ ] [In docs/14_ops_runbook.md around lines 24 to 30, the local build example](#in-docs14opsrunbookmd-around-lines-24-to-30-the-local-build-example)
- [ ] [In internal/config/config.go (around lines 154–160) and docs/14_ops_runbook.md](#in-internalconfigconfiggo-around-lines-154160-and-docs14opsrunbookmd)
- [ ] [In docs/api/dlq-remediation-pipeline.md around lines 168-176, the example](#in-docsapidlq-remediation-pipelinemd-around-lines-168-176-the-example)
- [ ] [In docs/api/dlq-remediation-pipeline.md around lines 313-319 (and also apply the](#in-docsapidlq-remediation-pipelinemd-around-lines-313-319-and-also-apply-the)
- [ ] [In docs/api/dlq-remediation-ui.md around lines 243 to 244, the example contains](#in-docsapidlq-remediation-uimd-around-lines-243-to-244-the-example-contains)
- [ ] [In docs/api/dlq-remediation-ui.md around lines 315-387 the API surface is](#in-docsapidlq-remediation-uimd-around-lines-315-387-the-api-surface-is)
- [ ] [In docs/SLAPS/FINAL-POSTMORTEM.md around line 20, the phrase "Zero](#in-docsslapsfinal-postmortemmd-around-line-20-the-phrase-zero)
- [ ] [In docs/SLAPS/FINAL-POSTMORTEM.md around lines 50 to 51, the claim "WCAG](#in-docsslapsfinal-postmortemmd-around-lines-50-to-51-the-claim-wcag)
- [ ] [In docs/SLAPS/FINAL-POSTMORTEM.md around lines 268 to 273, the resource metrics](#in-docsslapsfinal-postmortemmd-around-lines-268-to-273-the-resource-metrics)
- [ ] [In docs/SLAPS/FINAL-POSTMORTEM.md around lines 421 to 472, the Appendix contains](#in-docsslapsfinal-postmortemmd-around-lines-421-to-472-the-appendix-contains)
- [ ] [In docs/SLAPS/FINAL-POSTMORTEM.md around lines 433-434, the statement "Test](#in-docsslapsfinal-postmortemmd-around-lines-433-434-the-statement-test)
- [ ] [In README.md around lines 83-86 (and also lines 47-49 and 53-56), the examples](#in-readmemd-around-lines-83-86-and-also-lines-47-49-and-53-56-the-examples)
- [ ] [In README.md around lines 89 to 92, the example assumes bin/ exists so the go](#in-readmemd-around-lines-89-to-92-the-example-assumes-bin-exists-so-the-go)
- [ ] [In deployments/kubernetes/monitoring.yaml around lines 49-56, the alert expr](#in-deploymentskubernetesmonitoringyaml-around-lines-49-56-the-alert-expr)
- [ ] [In BUGS.md around line 41, the note warns that the current "short block per](#in-bugsmd-around-line-41-the-note-warns-that-the-current-short-block-per)
- [ ] [BUGS.md around lines 62-64: the comment asks to "wire exactly-once" but lacks](#bugsmd-around-lines-62-64-the-comment-asks-to-wire-exactly-once-but-lacks)
- [ ] [In BUGS.md around lines 65-66, the note currently lists BLMOVE as a](#in-bugsmd-around-lines-65-66-the-note-currently-lists-blmove-as-a)
- [ ] [In cmd/job-queue-system/main.go around lines 4-13 and also apply to lines 53-58,](#in-cmdjob-queue-systemmaingo-around-lines-4-13-and-also-apply-to-lines-53-58)
- [x] [In cmd/job-queue-system/main.go around lines 100 to 114, the signal handling](#in-cmdjob-queue-systemmaingo-around-lines-100-to-114-the-signal-handling)
- [ ] [In create_postmortem_tasks.py around lines 5, 18-19, 71-73, you are constructing](#in-createpostmortemtaskspy-around-lines-5-18-19-71-73-you-are-constructing)
- [ ] [In create_postmortem_tasks.py around lines 134 to 142, the JSON files are opened](#in-createpostmortemtaskspy-around-lines-134-to-142-the-json-files-are-opened)
- [ ] [In docs/api/advanced-rate-limiting-api.md around lines 74–84, the example shows](#in-docsapiadvanced-rate-limiting-apimd-around-lines-7484-the-example-shows)
- [ ] [In docs/api/advanced-rate-limiting-api.md around lines 140 to 158, the](#in-docsapiadvanced-rate-limiting-apimd-around-lines-140-to-158-the)
- [ ] [In docs/api/advanced-rate-limiting-api.md around lines 214–221 the field name](#in-docsapiadvanced-rate-limiting-apimd-around-lines-214221-the-field-name)
- [ ] [In docs/api/advanced-rate-limiting-api.md around lines 223 to 236, the Status](#in-docsapiadvanced-rate-limiting-apimd-around-lines-223-to-236-the-status)
- [ ] [In docs/api/advanced-rate-limiting-api.md around lines 365 to 375, the](#in-docsapiadvanced-rate-limiting-apimd-around-lines-365-to-375-the)
- [ ] [In docs/api/canary-deployments.md around lines 20 to 23, the docs state the](#in-docsapicanary-deploymentsmd-around-lines-20-to-23-the-docs-state-the)
- [ ] [In docs/api/canary-deployments.md around lines 40 to 49, the error codes table](#in-docsapicanary-deploymentsmd-around-lines-40-to-49-the-error-codes-table)
- [ ] [docs/api/canary-deployments.md lines 68-69 (and also update occurrences at](#docsapicanary-deploymentsmd-lines-68-69-and-also-update-occurrences-at)
- [ ] [In docs/api/canary-deployments.md around lines 100 to 116 (and also apply same](#in-docsapicanary-deploymentsmd-around-lines-100-to-116-and-also-apply-same)
- [ ] [In docs/api/canary-deployments.md around lines 110-113 (also apply fixes at](#in-docsapicanary-deploymentsmd-around-lines-110-113-also-apply-fixes-at)
- [ ] [In docs/api/canary-deployments.md around lines 146 to 162, the request body for](#in-docsapicanary-deploymentsmd-around-lines-146-to-162-the-request-body-for)
- [ ] [In docs/api/canary-deployments.md around lines 263 to 268, the throughput](#in-docsapicanary-deploymentsmd-around-lines-263-to-268-the-throughput)
- [ ] [In docs/api/canary-deployments.md around lines 761–786, update the WebSocket](#in-docsapicanary-deploymentsmd-around-lines-761786-update-the-websocket)
- [ ] [In docs/api/capacity-planning-api.md around lines 16 to 29, the CapacityPlanner](#in-docsapicapacity-planning-apimd-around-lines-16-to-29-the-capacityplanner)
- [ ] [In docs/api/capacity-planning-api.md around lines 289 to 308, the](#in-docsapicapacity-planning-apimd-around-lines-289-to-308-the)
- [ ] [In docs/api/capacity-planning-api.md around lines 365 to 373 the example calls](#in-docsapicapacity-planning-apimd-around-lines-365-to-373-the-example-calls)
- [ ] [In docs/api/capacity-planning-api.md around lines 479 to 494, the code](#in-docsapicapacity-planning-apimd-around-lines-479-to-494-the-code)
- [ ] [In docs/api/capacity-planning-api.md around lines 508 to 561, the error-handling](#in-docsapicapacity-planning-apimd-around-lines-508-to-561-the-error-handling)
- [ ] [In docs/api/capacity-planning-api.md around lines 508 to 533, add a canonical](#in-docsapicapacity-planning-apimd-around-lines-508-to-533-add-a-canonical)
- [ ] [In docs/api/capacity-planning-api.md around lines 631-640, the complexity table](#in-docsapicapacity-planning-apimd-around-lines-631-640-the-complexity-table)
- [ ] [In docs/SLAPS/worker-reflections/claude-001-reflection.md lines 1-4, the file](#in-docsslapsworker-reflectionsclaude-001-reflectionmd-lines-1-4-the-file)
- [ ] [In docs/SLAPS/worker-reflections/claude-005-reflection.md around lines 1 to 3,](#in-docsslapsworker-reflectionsclaude-005-reflectionmd-around-lines-1-to-3)
- [ ] [In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 24 to 27, the integration test](#in-eventhookstestdocumentationmd-around-lines-24-to-27-the-integration-test)
- [ ] [In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 218 to 225, the test commands](#in-eventhookstestdocumentationmd-around-lines-218-to-225-the-test-commands)
- [x] [In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 232 to 235, add an additional](#in-eventhookstestdocumentationmd-around-lines-232-to-235-add-an-additional)
- [ ] [In AGENTS.md around line 3, remove the trailing whitespace at the end of the](#in-agentsmd-around-line-3-remove-the-trailing-whitespace-at-the-end-of-the)
- [ ] [In AGENTS.md around line 10, the heading "Table of Contents" is currently an H1](#in-agentsmd-around-line-10-the-heading-table-of-contents-is-currently-an-h1)
- [ ] [In AGENTS.md around lines 11 to 39 the TOC list blocks do not have blank lines](#in-agentsmd-around-lines-11-to-39-the-toc-list-blocks-do-not-have-blank-lines)
- [ ] [In AGENTS.md around lines 34 to 36, the TOC contains a link text that includes](#in-agentsmd-around-lines-34-to-36-the-toc-contains-a-link-text-that-includes)
- [ ] [In AGENTS.md around lines 566 to 594, headings inside the blockquote/admonition](#in-agentsmd-around-lines-566-to-594-headings-inside-the-blockquoteadmonition)
- [ ] [In docs/api/admin-api.md around lines 10-11 (also apply same change at 41-46 and](#in-docsapiadmin-apimd-around-lines-10-11-also-apply-same-change-at-41-46-and)
- [ ] [In docs/api/admin-api.md around lines 82-84 (and also at 99-106 and 139-146),](#in-docsapiadmin-apimd-around-lines-82-84-and-also-at-99-106-and-139-146)
- [ ] [In docs/api/admin-api.md around lines 149 to 177 (and likewise apply to lines](#in-docsapiadmin-apimd-around-lines-149-to-177-and-likewise-apply-to-lines)
- [ ] [In docs/api/admin-api.md around lines 331 to 354, the Go example references a](#in-docsapiadmin-apimd-around-lines-331-to-354-the-go-example-references-a)
- [ ] [`](#)
- [ ] [.claude/agents/feature-enhancer.md around lines 38 to 42: the file uses a](#claudeagentsfeature-enhancermd-around-lines-38-to-42-the-file-uses-a)
- [ ] [In .claude/agents/feature-enhancer.md around lines 165 to 190, replace the](#in-claudeagentsfeature-enhancermd-around-lines-165-to-190-replace-the)
- [ ] [In .githooks/pre-commit around lines 19 to 25, the script unconditionally runs](#in-githookspre-commit-around-lines-19-to-25-the-script-unconditionally-runs)
- [x] [In .gitignore around line 16, you added / .gocache/ but missed other common](#in-gitignore-around-line-16-you-added--gocache-but-missed-other-common)
- [ ] [AGENTS.md lines 828-830: the second "APPENDIX B: WILD IDEAS — HAVE A BRAINSTORM"](#agentsmd-lines-828-830-the-second-appendix-b-wild-ideas--have-a-brainstorm)
- [ ] [In BUGS.md around lines 3-4 (also apply the same change to lines 51-53 and](#in-bugsmd-around-lines-3-4-also-apply-the-same-change-to-lines-51-53-and)
- [ ] [In BUGS.md around lines 12–47: the heartbeat example blocks the worker (no](#in-bugsmd-around-lines-1247-the-heartbeat-example-blocks-the-worker-no)
- [ ] [In BUGS.md around lines 55 to 69 (and also line 71) the worker registry](#in-bugsmd-around-lines-55-to-69-and-also-line-71-the-worker-registry)
- [ ] [In BUGS.md around lines 119–121, the ledger guidance needs concrete failure-mode](#in-bugsmd-around-lines-119121-the-ledger-guidance-needs-concrete-failure-mode)
- [ ] [In cmd/admin-api/main.go lines 80-89, the missing-config check incorrectly uses](#in-cmdadmin-apimaingo-lines-80-89-the-missing-config-check-incorrectly-uses)
- [ ] [In cmd/tui/main.go around lines 38 to 55, don't discard flag usage output and](#in-cmdtuimaingo-around-lines-38-to-55-dont-discard-flag-usage-output-and)
- [ ] [In cmd/tui/main.go around lines 74 to 94, the current logic discards the host](#in-cmdtuimaingo-around-lines-74-to-94-the-current-logic-discards-the-host)
- [ ] [In cmd/tui/main.go around lines 109 to 112, the Redis Ping is using the](#in-cmdtuimaingo-around-lines-109-to-112-the-redis-ping-is-using-the)
- [ ] [In dependency_analysis.py around lines 283 to 287 (and also the earlier mapping](#in-dependencyanalysispy-around-lines-283-to-287-and-also-the-earlier-mapping)
- [ ] [In dependency_analysis.py around lines 312 to 324, the function](#in-dependencyanalysispy-around-lines-312-to-324-the-function)
- [ ] [In deployments/admin-api/docker-compose.yaml around lines 35-36, the config](#in-deploymentsadmin-apidocker-composeyaml-around-lines-35-36-the-config)
- [ ] [In deployments/admin-api/k8s-deployment.yaml around lines 41 to 44, the inline](#in-deploymentsadmin-apik8s-deploymentyaml-around-lines-41-to-44-the-inline)
- [ ] [In deployments/docker/admin-api.env.example lines 1-4, replace the toy token](#in-deploymentsdockeradmin-apienvexample-lines-1-4-replace-the-toy-token)
- [x] [In deployments/docker/Dockerfile.admin-api around lines 52 to 53, the](#in-deploymentsdockerdockerfileadmin-api-around-lines-52-to-53-the)
- [ ] [In deployments/kubernetes/admin-api-deployment.yaml around lines 114 to 123 (and](#in-deploymentskubernetesadmin-api-deploymentyaml-around-lines-114-to-123-and)
- [ ] [In deployments/kubernetes/admin-api-deployment.yaml around lines 134 to 147, the](#in-deploymentskubernetesadmin-api-deploymentyaml-around-lines-134-to-147-the)
- [ ] [deployments/scripts/lib/logging.sh lines 4-6: the guard pattern that](#deploymentsscriptslibloggingsh-lines-4-6-the-guard-pattern-that)
- [ ] [In deployments/scripts/test-staging-deployment.sh around lines 278 to 287,](#in-deploymentsscriptstest-staging-deploymentsh-around-lines-278-to-287)
- [ ] [In deployments/scripts/test-staging-deployment.sh around lines 336 to 346,](#in-deploymentsscriptstest-staging-deploymentsh-around-lines-336-to-346)
- [ ] [In docs/api/_index.md around lines 7–13, the versioning policy is vague; update](#in-docsapiindexmd-around-lines-713-the-versioning-policy-is-vague-update)
- [ ] [In docs/api/calendar-view.md around lines 84 to 101, the example and any](#in-docsapicalendar-viewmd-around-lines-84-to-101-the-example-and-any)
- [ ] [In docs/api/calendar-view.md around lines 465 to 491, the example response leaks](#in-docsapicalendar-viewmd-around-lines-465-to-491-the-example-response-leaks)
- [ ] [In docs/api/calendar-view.md around lines 519 to 548, the error examples and](#in-docsapicalendar-viewmd-around-lines-519-to-548-the-error-examples-and)
- [ ] [In docs/api/calendar-view.md around lines 739 to 758, the authentication](#in-docsapicalendar-viewmd-around-lines-739-to-758-the-authentication)
- [ ] [In docs/SLAPS/coordinator-observations.md around lines 195 to 197, fix the](#in-docsslapscoordinator-observationsmd-around-lines-195-to-197-fix-the)
- [ ] [In docs/SLAPS/coordinator-observations.md around lines 235 to 236, the "Total](#in-docsslapscoordinator-observationsmd-around-lines-235-to-236-the-total)
- [ ] [In docs/SLAPS/worker-reflections/claude-001-reflection.md around lines 39 to 41](#in-docsslapsworker-reflectionsclaude-001-reflectionmd-around-lines-39-to-41)
- [ ] [In docs/YOU ARE WORKER 6.md around lines 8–16, the current claim protocol is](#in-docsyou-are-worker-6md-around-lines-816-the-current-claim-protocol-is)
- [ ] [In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 248 to 255, the perf table](#in-eventhookstestdocumentationmd-around-lines-248-to-255-the-perf-table)
- [ ] [In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 318 to 321, the claim that](#in-eventhookstestdocumentationmd-around-lines-318-to-321-the-claim-that)
- [ ] [In .github/workflows/update-progress.yml around lines 30 to 33, the workflow](#in-githubworkflowsupdate-progressyml-around-lines-30-to-33-the-workflow)
- [ ] [In create_review_tasks.py around lines 107-112 (and also line 141), the test](#in-createreviewtaskspy-around-lines-107-112-and-also-line-141-the-test)
- [x] [In deployments/admin-api/k8s-deployment.yaml around lines 71 to 116, the](#in-deploymentsadmin-apik8s-deploymentyaml-around-lines-71-to-116-the)
- [ ] [In deployments/admin-api/k8s-deployment.yaml around lines 152–167, the](#in-deploymentsadmin-apik8s-deploymentyaml-around-lines-152167-the)
- [ ] [In deployments/docker/grafana/datasources/prometheus.yaml around lines 3 to 9,](#in-deploymentsdockergrafanadatasourcesprometheusyaml-around-lines-3-to-9)
- [ ] [In deployments/kubernetes/rbac-monitoring.yaml around lines 35 to 42, the alert](#in-deploymentskubernetesrbac-monitoringyaml-around-lines-35-to-42-the-alert)
- [ ] [In deployments/kubernetes/rbac-monitoring.yaml around lines 45 to 51 (and also](#in-deploymentskubernetesrbac-monitoringyaml-around-lines-45-to-51-and-also)
- [ ] [In deployments/kubernetes/rbac-monitoring.yaml around lines 161–173, the two](#in-deploymentskubernetesrbac-monitoringyaml-around-lines-161173-the-two)
- [ ] [In deployments/kubernetes/rbac-token-service-deployment.yaml around lines 367 to](#in-deploymentskubernetesrbac-token-service-deploymentyaml-around-lines-367-to)
- [ ] [In deployments/scripts/deploy-rbac-staging.sh around line 11, ShellCheck is](#in-deploymentsscriptsdeploy-rbac-stagingsh-around-line-11-shellcheck-is)
- [ ] [In deployments/scripts/deploy-rbac-staging.sh around lines 13 to 35, the](#in-deploymentsscriptsdeploy-rbac-stagingsh-around-lines-13-to-35-the)
- [ ] [In deployments/scripts/deploy-rbac-staging.sh around lines 36 to 47, the build](#in-deploymentsscriptsdeploy-rbac-stagingsh-around-lines-36-to-47-the-build)
- [ ] [In deployments/scripts/deploy-rbac-staging.sh around lines 65 to 68, the script](#in-deploymentsscriptsdeploy-rbac-stagingsh-around-lines-65-to-68-the-script)
- [ ] [In deployments/scripts/deploy-staging.sh around lines 184-185 you call](#in-deploymentsscriptsdeploy-stagingsh-around-lines-184-185-you-call)
- [ ] [In deployments/scripts/setup-monitoring.sh around line 11, the script blindly](#in-deploymentsscriptssetup-monitoringsh-around-line-11-the-script-blindly)
- [ ] [In deployments/scripts/setup-monitoring.sh around lines 103 to 108, the scrape](#in-deploymentsscriptssetup-monitoringsh-around-lines-103-to-108-the-scrape)
- [ ] [In deployments/scripts/setup-monitoring.sh around lines 146 to 218, the script](#in-deploymentsscriptssetup-monitoringsh-around-lines-146-to-218-the-script)
- [ ] [In deployments/scripts/setup-monitoring.sh around lines 211 to 215, the secret](#in-deploymentsscriptssetup-monitoringsh-around-lines-211-to-215-the-secret)
- [ ] [In docs/07_test_plan.md around lines 53 to 58, the CI/workflow and docs are out](#in-docs07testplanmd-around-lines-53-to-58-the-ciworkflow-and-docs-are-out)
- [ ] [In docs/12_performance_baseline.md around lines 46 to 53, the instruction to set](#in-docs12performancebaselinemd-around-lines-46-to-53-the-instruction-to-set)
- [ ] [In docs/api/anomaly-radar-openapi.yaml around lines 80-109 (and similarly](#in-docsapianomaly-radar-openapiyaml-around-lines-80-109-and-similarly)
- [ ] [In docs/api/anomaly-radar-openapi.yaml around lines 229 to 236 (and similarly](#in-docsapianomaly-radar-openapiyaml-around-lines-229-to-236-and-similarly)
- [ ] [In docs/api/anomaly-radar-openapi.yaml around lines 386 to 393, the description](#in-docsapianomaly-radar-openapiyaml-around-lines-386-to-393-the-description)
- [ ] [In docs/api/anomaly-radar-openapi.yaml around lines 433-446, numeric fields lack](#in-docsapianomaly-radar-openapiyaml-around-lines-433-446-numeric-fields-lack)
- [ ] [docs/api/anomaly-radar-slo-budget.md around lines 223 to 241: the /status](#docsapianomaly-radar-slo-budgetmd-around-lines-223-to-241-the-status)
- [ ] [docs/api/canary-deployments.md around lines 68-76 (and also apply to 366-373,](#docsapicanary-deploymentsmd-around-lines-68-76-and-also-apply-to-366-373)
- [ ] [In docs/api/canary-deployments.md around lines 84-90 (and likewise at 112-118,](#in-docsapicanary-deploymentsmd-around-lines-84-90-and-likewise-at-112-118)
- [ ] [In docs/api/canary-deployments.md around lines 90-92 (and also update lines](#in-docsapicanary-deploymentsmd-around-lines-90-92-and-also-update-lines)
- [ ] [In docs/api/canary-deployments.md around lines 90-92 (and also adjust at lines](#in-docsapicanary-deploymentsmd-around-lines-90-92-and-also-adjust-at-lines)
- [ ] [In docs/api/canary-deployments.md around lines 304-316 (and also adjust](#in-docsapicanary-deploymentsmd-around-lines-304-316-and-also-adjust)
- [ ] [In docs/api/canary-deployments.md around lines 732-739 (and similarly lines](#in-docsapicanary-deploymentsmd-around-lines-732-739-and-similarly-lines)
- [ ] [In docs/api/capacity-planning-api.md around lines 317 to 324, the example](#in-docsapicapacity-planning-apimd-around-lines-317-to-324-the-example)
- [ ] [In docs/TUI/README.md around lines 31 to 33, replace the nonstandard admonition](#in-docstuireadmemd-around-lines-31-to-33-replace-the-nonstandard-admonition)
- [ ] [In .github/workflows/update-progress.yml around lines 34 to 54, the commit step](#in-githubworkflowsupdate-progressyml-around-lines-34-to-54-the-commit-step)
- [ ] [In .github/workflows/update-progress.yml around lines 56-58, the push step](#in-githubworkflowsupdate-progressyml-around-lines-56-58-the-push-step)
- [ ] [In deployments/admin-api/Dockerfile around lines 18 to 21, the Go build step](#in-deploymentsadmin-apidockerfile-around-lines-18-to-21-the-go-build-step)
- [ ] [In deployments/admin-api/Dockerfile around lines 19 to 21, the ldflags usage -X](#in-deploymentsadmin-apidockerfile-around-lines-19-to-21-the-ldflags-usage--x)
- [ ] [deployments/admin-api/Dockerfile lines 23-46: the image lacks standard OCI](#deploymentsadmin-apidockerfile-lines-23-46-the-image-lacks-standard-oci)
- [ ] [In deployments/admin-api/Dockerfile around lines 49 to 50, remove the Docker](#in-deploymentsadmin-apidockerfile-around-lines-49-to-50-remove-the-docker)
- [ ] [In AGENTS.md around lines 791 to 828, there is a duplicated "## APPENDIX B: WILD](#in-agentsmd-around-lines-791-to-828-there-is-a-duplicated--appendix-b-wild)
- [ ] [In append_metadata.py around lines 53 to 80, the function currently appends YAML](#in-appendmetadatapy-around-lines-53-to-80-the-function-currently-appends-yaml)
- [ ] [In append_metadata.py around lines 184-194, the DAG write is fine but the file](#in-appendmetadatapy-around-lines-184-194-the-dag-write-is-fine-but-the-file)
- [ ] [In BUGS.md around lines 1-7 (and similarly lines 131-136), the tone is informal](#in-bugsmd-around-lines-1-7-and-similarly-lines-131-136-the-tone-is-informal)
- [ ] [In BUGS.md around lines 14-21, the current code treats SetArgs as returning](#in-bugsmd-around-lines-14-21-the-current-code-treats-setargs-as-returning)
- [x] [In BUGS.md around lines 51 to 53, update the note about worker registry to](#in-bugsmd-around-lines-51-to-53-update-the-note-about-worker-registry-to)
- [ ] [In BUGS.md around lines 81-115, the current mover pops entries with ZPOPMIN then](#in-bugsmd-around-lines-81-115-the-current-mover-pops-entries-with-zpopmin-then)
- [ ] [In cmd/admin-api/main.go around lines 32 to 35, the code checks the error return](#in-cmdadmin-apimaingo-around-lines-32-to-35-the-code-checks-the-error-return)
- [ ] [In cmd/admin-api/main.go around line 59, the deferred call defer logger.Sync()](#in-cmdadmin-apimaingo-around-line-59-the-deferred-call-defer-loggersync)
- [ ] [In cmd/admin-api/main.go around line 67 (and similarly for lines 98-112), the](#in-cmdadmin-apimaingo-around-line-67-and-similarly-for-lines-98-112-the)
- [ ] [In cmd/admin-api/main.go around lines 69 to 71, remove the call to logger.Fatal](#in-cmdadmin-apimaingo-around-lines-69-to-71-remove-the-call-to-loggerfatal)
- [ ] [In cmd/admin-api/main.go around lines 84 to 89, the current error check uses](#in-cmdadmin-apimaingo-around-lines-84-to-89-the-current-error-check-uses)
- [ ] [In create_review_tasks.py around lines 100-113 (and also update the duplicate](#in-createreviewtaskspy-around-lines-100-113-and-also-update-the-duplicate)
- [x] [In demos/responsive-tui.tape around lines 19-27, the script sets COLUMNS](#in-demosresponsive-tuitape-around-lines-19-27-the-script-sets-columns)
- [ ] [In demos/responsive-tui.tape around lines 73-74 (also apply same fix at 131-132,](#in-demosresponsive-tuitape-around-lines-73-74-also-apply-same-fix-at-131-132)
- [ ] [In dependency_analysis.py around lines 302 to 324, the "provides" entries are](#in-dependencyanalysispy-around-lines-302-to-324-the-provides-entries-are)
- [ ] [In deployments/scripts/test-staging-deployment.sh around lines 441 to 449, the](#in-deploymentsscriptstest-staging-deploymentsh-around-lines-441-to-449-the)
- [ ] [docs/07_test_plan.md lines 28–37: the chaos tests as written require](#docs07testplanmd-lines-2837-the-chaos-tests-as-written-require)
- [ ] [In docs/07_test_plan.md around lines 53 to 57, the CPU governor step assumes](#in-docs07testplanmd-around-lines-53-to-57-the-cpu-governor-step-assumes)
- [ ] [In docs/15_promotion_checklists.md around line 3, the "Last updated" timestamp](#in-docs15promotionchecklistsmd-around-line-3-the-last-updated-timestamp)
- [ ] [In docs/api/calendar-view.md around lines 63 to 77, the RecurringRule sample](#in-docsapicalendar-viewmd-around-lines-63-to-77-the-recurringrule-sample)
- [ ] [In docs/api/calendar-view.md around lines 198 to 206, the request example](#in-docsapicalendar-viewmd-around-lines-198-to-206-the-request-example)
- [ ] [In docs/api/canary-deployments.md around lines 558 to 597, replace all example](#in-docsapicanary-deploymentsmd-around-lines-558-to-597-replace-all-example)
- [ ] [In docs/api/chaos-harness.md around lines 430 to 441, document the semantics of](#in-docsapichaos-harnessmd-around-lines-430-to-441-document-the-semantics-of)
- [ ] [In docs/api/chaos-harness.md around lines 601 to 626, the document lacks a](#in-docsapichaos-harnessmd-around-lines-601-to-626-the-document-lacks-a)
- [ ] [In docs/PRD.md around lines 119 to 124, the heartbeat key contains a stray](#in-docsprdmd-around-lines-119-to-124-the-heartbeat-key-contains-a-stray)
- [ ] [In docs/SLAPS/worker-reflections/claude-001-reflection.md around lines 111 to](#in-docsslapsworker-reflectionsclaude-001-reflectionmd-around-lines-111-to)
- [ ] [In docs/SLAPS/worker-reflections/claude-008-reflection.md around lines 1 to 4,](#in-docsslapsworker-reflectionsclaude-008-reflectionmd-around-lines-1-to-4)
- [ ] [docs/SLAPS/worker-reflections/claude-008-reflection.md around lines 27-33:](#docsslapsworker-reflectionsclaude-008-reflectionmd-around-lines-27-33)
- [ ] [docs/SLAPS/worker-reflections/claude-008-reflection.md lines 37-41: the strategy](#docsslapsworker-reflectionsclaude-008-reflectionmd-lines-37-41-the-strategy)
- [ ] [In docs/SLAPS/worker-reflections/claude-008-reflection.md around lines 48–51,](#in-docsslapsworker-reflectionsclaude-008-reflectionmd-around-lines-4851)
- [ ] [In docs/YOU ARE WORKER 6.md around lines 8-9 (and also lines 27-29), the](#in-docsyou-are-worker-6md-around-lines-8-9-and-also-lines-27-29-the)
- [ ] [In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 20 to 26, the documented test](#in-eventhookstestdocumentationmd-around-lines-20-to-26-the-documented-test)
- [ ] [In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 246 to 265, the perf tables](#in-eventhookstestdocumentationmd-around-lines-246-to-265-the-perf-tables)
- [ ] [In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 250–255 (and also update lines](#in-eventhookstestdocumentationmd-around-lines-250255-and-also-update-lines)
- [ ] [In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 252 to 256, the note claiming](#in-eventhookstestdocumentationmd-around-lines-252-to-256-the-note-claiming)
- [ ] [In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 331 to 355, the example test](#in-eventhookstestdocumentationmd-around-lines-331-to-355-the-example-test)
- [ ] [In .claude/agents/feature-enhancer.md lines 1-6, the file triggers markdownlint](#in-claudeagentsfeature-enhancermd-lines-1-6-the-file-triggers-markdownlint)
- [ ] [In cmd/job-queue-system/main.go around lines 182 to 189, the command prints a](#in-cmdjob-queue-systemmaingo-around-lines-182-to-189-the-command-prints-a)
- [ ] [In deployments/admin-api/deploy.sh around lines 101-107 you currently print a](#in-deploymentsadmin-apideploysh-around-lines-101-107-you-currently-print-a)
- [ ] [In deployments/docker/docker-compose.yaml around lines 20 to 56 (and also apply](#in-deploymentsdockerdocker-composeyaml-around-lines-20-to-56-and-also-apply)
- [ ] [In deployments/docker/rbac-configs/roles.yaml around lines 107 to 116, the RBAC](#in-deploymentsdockerrbac-configsrolesyaml-around-lines-107-to-116-the-rbac)
- [x] [In deployments/kubernetes/rbac-monitoring.yaml around lines 33 to 43, the alert](#in-deploymentskubernetesrbac-monitoringyaml-around-lines-33-to-43-the-alert)
- [ ] [In deployments/kubernetes/rbac-monitoring.yaml around lines 45 to 59, the alert](#in-deploymentskubernetesrbac-monitoringyaml-around-lines-45-to-59-the-alert)
- [ ] [In deployments/kubernetes/rbac-token-service-deployment.yaml around lines](#in-deploymentskubernetesrbac-token-service-deploymentyaml-around-lines)
- [ ] [In deployments/kubernetes/rbac-token-service-deployment.yaml lines 199-201 (and](#in-deploymentskubernetesrbac-token-service-deploymentyaml-lines-199-201-and)
- [ ] [In deployments/scripts/deploy-rbac-staging.sh around line 11, the dynamic source](#in-deploymentsscriptsdeploy-rbac-stagingsh-around-line-11-the-dynamic-source)
- [ ] [In deployments/scripts/deploy-rbac-staging.sh around lines 40-47 (and also apply](#in-deploymentsscriptsdeploy-rbac-stagingsh-around-lines-40-47-and-also-apply)
- [ ] [In deployments/scripts/deploy-staging.sh around lines 52 to 64, the file defines](#in-deploymentsscriptsdeploy-stagingsh-around-lines-52-to-64-the-file-defines)
- [ ] [In deployments/scripts/deploy-staging.sh around lines 100 to 118, the IMAGE_NAME](#in-deploymentsscriptsdeploy-stagingsh-around-lines-100-to-118-the-imagename)
- [ ] [In deployments/scripts/deploy-staging.sh around lines 147-155, the script uses a](#in-deploymentsscriptsdeploy-stagingsh-around-lines-147-155-the-script-uses-a)
- [ ] [In deployments/scripts/deploy-staging.sh around lines 182 to 186 there is a](#in-deploymentsscriptsdeploy-stagingsh-around-lines-182-to-186-there-is-a)
- [ ] [In deployments/scripts/health-check-rbac.sh around lines 41 to 44, the kubectl](#in-deploymentsscriptshealth-check-rbacsh-around-lines-41-to-44-the-kubectl)
- [ ] [In deployments/scripts/health-check-rbac.sh around lines 354 to 380, the parsed](#in-deploymentsscriptshealth-check-rbacsh-around-lines-354-to-380-the-parsed)
- [ ] [In deployments/scripts/setup-monitoring.sh around line 11, ShellCheck is](#in-deploymentsscriptssetup-monitoringsh-around-line-11-shellcheck-is)
- [x] [In deployments/scripts/setup-monitoring.sh around lines 30-31 (and the similar](#in-deploymentsscriptssetup-monitoringsh-around-lines-30-31-and-the-similar)
- [ ] [In deployments/scripts/setup-monitoring.sh around lines 117 to 120, the](#in-deploymentsscriptssetup-monitoringsh-around-lines-117-to-120-the)
- [ ] [In deployments/scripts/setup-monitoring.sh around lines 146–214, the script](#in-deploymentsscriptssetup-monitoringsh-around-lines-146214-the-script)
- [ ] [In deployments/scripts/test-staging-deployment.sh around lines 278 to 280,](#in-deploymentsscriptstest-staging-deploymentsh-around-lines-278-to-280)
- [ ] [In deployments/scripts/test-staging-deployment.sh around lines 281 to 314, the](#in-deploymentsscriptstest-staging-deploymentsh-around-lines-281-to-314-the)
- [ ] [In deployments/scripts/test-staging-deployment.sh around lines 321 to 376, the](#in-deploymentsscriptstest-staging-deploymentsh-around-lines-321-to-376-the)
- [ ] [In docs/12_performance_baseline.md around line 3, the "Last updated" stamp is](#in-docs12performancebaselinemd-around-line-3-the-last-updated-stamp-is)
- [ ] [In docs/12_performance_baseline.md around lines 51-53, the doc runs](#in-docs12performancebaselinemd-around-lines-51-53-the-doc-runs)
- [ ] [In docs/12_performance_baseline.md around lines 64-65, replace the vague "curl](#in-docs12performancebaselinemd-around-lines-64-65-replace-the-vague-curl)
- [ ] [In docs/api/admin-api.md around lines 360 to 366, the health endpoint is](#in-docsapiadmin-apimd-around-lines-360-to-366-the-health-endpoint-is)
- [ ] [In docs/api/capacity-planning-api.md around lines 322-324, the example import](#in-docsapicapacity-planning-apimd-around-lines-322-324-the-example-import)
- [ ] [In docs/FEATURE_ENHANCEMENT_AGENT_PROMPT.md around lines 156 to 161 the link](#in-docsfeatureenhancementagentpromptmd-around-lines-156-to-161-the-link)
- [ ] [In docs/PRD.md around lines 154–156, the phrase "scan" implies using Redis KEYS](#in-docsprdmd-around-lines-154156-the-phrase-scan-implies-using-redis-keys)
- [ ] [In Makefile around line 6: the LDFLAGS uses -X main.version=$(VERSION) but there](#in-makefile-around-line-6-the-ldflags-uses--x-mainversionversion-but-there)
- [ ] [In Makefile around lines 22 to 24, the run-tui target hardcodes "./bin/tui"](#in-makefile-around-lines-22-to-24-the-run-tui-target-hardcodes-bintui)
- [ ] [In create_postmortem_tasks.py around lines 18-19 (and also apply the same change](#in-createpostmortemtaskspy-around-lines-18-19-and-also-apply-the-same-change)
- [ ] [In create_postmortem_tasks.py around lines 132 to 141 the code writes JSON files](#in-createpostmortemtaskspy-around-lines-132-to-141-the-code-writes-json-files)
- [ ] [In create_postmortem_tasks.py around lines 134-141 (and similarly lines](#in-createpostmortemtaskspy-around-lines-134-141-and-similarly-lines)
- [ ] [In deployments/admin-api/k8s-deployment.yaml around lines 41 to 44, the inline](#in-deploymentsadmin-apik8s-deploymentyaml-around-lines-41-to-44-the-inline)
- [ ] [In deployments/README.md around lines 186 to 190 there is a duplicated bullet](#in-deploymentsreadmemd-around-lines-186-to-190-there-is-a-duplicated-bullet)
- [ ] [In docs/12_performance_baseline.md around lines 26 to 29, the Redis options (AOF](#in-docs12performancebaselinemd-around-lines-26-to-29-the-redis-options-aof)
- [ ] [In docs/12_performance_baseline.md around lines 58 to 62, the example uses](#in-docs12performancebaselinemd-around-lines-58-to-62-the-example-uses)
- [ ] [In docs/12_performance_baseline.md around lines 76 to 77, the expected-results](#in-docs12performancebaselinemd-around-lines-76-to-77-the-expected-results)
- [ ] [In docs/api/anomaly-radar-openapi.yaml around lines 32-35, the inline mapping](#in-docsapianomaly-radar-openapiyaml-around-lines-32-35-the-inline-mapping)
- [ ] [In docs/api/anomaly-radar-openapi.yaml around lines 80 to 101 (and also apply](#in-docsapianomaly-radar-openapiyaml-around-lines-80-to-101-and-also-apply)
- [ ] [In docs/api/anomaly-radar-openapi.yaml around lines 229-236 (and also update the](#in-docsapianomaly-radar-openapiyaml-around-lines-229-236-and-also-update-the)
- [ ] [In docs/api/anomaly-radar-openapi.yaml around lines 404-412, the 'metrics' array](#in-docsapianomaly-radar-openapiyaml-around-lines-404-412-the-metrics-array)
- [ ] [In docs/api/anomaly-radar-slo-budget.md around lines 262 to 295 the example GET](#in-docsapianomaly-radar-slo-budgetmd-around-lines-262-to-295-the-example-get)
- [ ] [In docs/api/anomaly-radar-slo-budget.md around lines 582-592 (also apply changes](#in-docsapianomaly-radar-slo-budgetmd-around-lines-582-592-also-apply-changes)
- [ ] [In README.md around lines 201 to 203, the docker compose command references the](#in-readmemd-around-lines-201-to-203-the-docker-compose-command-references-the)

## In .github/CODEOWNERS around lines 5 to 18, the file duplicates ownership

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567117

- [review_comment] 2025-09-16T03:15:00Z by coderabbitai[bot] (.github/CODEOWNERS:18)

```text
In .github/CODEOWNERS around lines 5 to 18, the file duplicates ownership
entries that are already covered by the top-level wildcard "* @flyingrobots";
remove the redundant pattern lines (.github/**, deploy/**, cmd/**, internal/**,
test/**) so that the single wildcard entry remains as the default owner, keeping
only any specific overrides if truly needed.
```

> [!INFO]- **Rejected**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | No | Yes | 10 | Comment payload contained only a stray HTML closing tag; nothing actionable. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Reviewer note deemed non-actionable. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:133
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## </summary>

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567117

- [review_comment] 2025-09-16T03:15:00Z by coderabbitai[bot] (.github/CODEOWNERS:18)

```text
</summary>
```

> [!INFO]- **Rejected**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | No | Yes | 10 | Comment payload contained only a stray HTML closing tag; nothing actionable. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Reviewer note deemed non-actionable. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:133
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## .github/workflows/changelog.yml around lines 28-29: the command uses unquoted

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567123

- [review_comment] 2025-09-16T03:15:00Z by coderabbitai[bot] (.github/workflows/changelog.yml:29)

```text
.github/workflows/changelog.yml around lines 28-29: the command uses unquoted
command substitution $(go env GOPATH) which can break due to word-splitting
(SC2046); change it to quote the substitution so the path is treated as a single
word (e.g. use "$(go env GOPATH)/bin/git-chglog" in the run line) and keep the
existing fallback (|| echo ...) as-is so the workflow behaves the same on
failure.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Confirmed the workflow already wraps the substitution in quotes and preserved fallback behaviour. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Confirmed the workflow already wraps the substitution in quotes and preserved fallback behaviour. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:168
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In .github/workflows/changelog.yml around lines 36-38, the workflow pushes

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567131

- [review_comment] 2025-09-16T03:15:01Z by coderabbitai[bot] (.github/workflows/changelog.yml:42)

```text
In .github/workflows/changelog.yml around lines 36-38, the workflow pushes
directly to the default branch and can have concurrent runs collide; add a
top-level concurrency stanza to serialize runs (use a stable group key tied to
the repository/workflow, e.g. "changelog-${{ github.repository }}-${{
github.workflow }}" or include default_branch) and set cancel-in-progress: false
so only one push to the default branch runs at a time; add this concurrency
block at the top of the workflow YAML to prevent race-prone pushes.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## .github/workflows/ci.yml around lines 38 to 45: the workflow immediately starts

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567136

- [review_comment] 2025-09-16T03:15:01Z by coderabbitai[bot] (.github/workflows/ci.yml:45)

```text
.github/workflows/ci.yml around lines 38 to 45: the workflow immediately starts
5 E2E runs against localhost:6379 without verifying Redis is ready, causing
flaky failures; add a readiness gate prior to the for-loop that waits for Redis
to respond (e.g., loop using redis-cli ping until it returns PONG or timeout, or
use a small curl/openssl TCP probe) and fail fast if timeout is reached; place
the wait logic just before the for i in {1..5} loop so tests only start after
Redis health is confirmed.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | CI workflow now waits for Redis before the E2E loop, matching the requested gating. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> CI workflow now waits for Redis before the E2E loop, matching the requested gating. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:241
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In .github/workflows/goreleaser.yml around lines 25 to 27, the echo redirections

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567141

- [review_comment] 2025-09-16T03:15:01Z by coderabbitai[bot] (.github/workflows/goreleaser.yml:27)

```text
In .github/workflows/goreleaser.yml around lines 25 to 27, the echo redirections
use an unquoted $GITHUB_ENV which can break on filenames with spaces or special
chars; update the two lines so the redirection target is quoted (use >>
"$GITHUB_ENV") and keep the echoed strings quoted as-is to ensure safe, portable
assignment into the env file.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Updated workflow writes to `"$GITHUB_ENV"`, eliminating the unquoted redirection. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Updated workflow writes to `"$GITHUB_ENV"`, eliminating the unquoted redirection. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:278
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In .github/workflows/goreleaser.yml around lines 28-33, the workflow logs into

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567149

- [review_comment] 2025-09-16T03:15:01Z by coderabbitai[bot] (.github/workflows/goreleaser.yml:37)

```text
In .github/workflows/goreleaser.yml around lines 28-33, the workflow logs into
GHCR but does not set up QEMU or Docker Buildx for multi-arch builds; add steps
before the login/build steps to (1) register QEMU emulators (use
actions/setup-qemu-action@v2) and (2) create/enable a buildx builder (use
docker/setup-buildx-action@v2), ensuring buildx is the active builder and
supports the target platforms; keep the login step but then invoke buildx-based
multi-platform build/push (or ensure goreleaser step uses buildx) so multi-arch
images are built correctly.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## .github/workflows/markdownlint.yml around line 6: the branch array uses spaces

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567155

- [review_comment] 2025-09-16T03:15:01Z by coderabbitai[bot] (.github/workflows/markdownlint.yml:6)

```text
.github/workflows/markdownlint.yml around line 6: the branch array uses spaces
inside the brackets ("[ main ]") which trips YAML/style linters; change it to
use no inner spacing ("[main]") so the array is formatted as a compact literal
and passes linting.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Branch array rewritten as `[main]`, satisfying markdownlint and YAML format guidance. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Branch array rewritten as `[main]`, satisfying markdownlint and YAML format guidance. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:351
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## .github/workflows/markdownlint.yml lines 12-21: the workflow lacks

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567164

- [review_comment] 2025-09-16T03:15:02Z by coderabbitai[bot] (.github/workflows/markdownlint.yml:21)

```text
.github/workflows/markdownlint.yml lines 12-21: the workflow lacks
least-privilege permissions and concurrency control; add a permissions block
granting only what the job needs (e.g., permissions: contents: read) at the
workflow level and add a concurrency key to cancel duplicate runs (e.g., group
using workflow/ref or workflow/run id with cancel-in-progress: true) so runners
aren’t wasted and attack surface is reduced.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Workflow now ships with least-privilege permissions and concurrency guard as requested. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Workflow now ships with least-privilege permissions and concurrency guard as requested. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:385
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## .github/workflows/markdownlint.yml lines 12–16: the workflow uses mutable tags

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567169

- [review_comment] 2025-09-16T03:15:02Z by coderabbitai[bot] (.github/workflows/markdownlint.yml:16)

```text
.github/workflows/markdownlint.yml lines 12–16: the workflow uses mutable tags
for actions; update the two uses entries to pinned commit SHAs as suggested —
change actions/checkout@v4 to
actions/checkout@08eba0b27e820071cde6df949e0beb9ba4906955 (keep with:
fetch-depth: 0) and change DavidAnson/markdownlint-cli2-action@v17 to
DavidAnson/markdownlint-cli2-action@db43aef879112c3119a410d69f66701e0d530809 so
both actions are pinned to a specific commit SHA.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Workflow now pins `actions/checkout` and `markdownlint-cli2-action` to commit SHAs. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Workflow now pins `actions/checkout` and `markdownlint-cli2-action` to commit SHAs. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:421
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## .goreleaser.yaml around lines 15 to 20: the archives block currently produces

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567181

- [review_comment] 2025-09-16T03:15:02Z by coderabbitai[bot] (.goreleaser.yaml:20)

```text
.goreleaser.yaml around lines 15 to 20: the archives block currently produces
tar.gz for all OSes (including Windows); change it to keep tar.gz for
non-Windows and add a format_overrides entry that sets format: zip for
goos/windows so Windows builds produce zip archives. Update the archives stanza
to include format_overrides with a selector for goos: windows -> format: zip
(and ensure name_template remains appropriate for both formats).
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Added Windows-specific format override so GoReleaser emits `.zip` archives. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added Windows-specific format override so GoReleaser emits `.zip` archives. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:458
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## .goreleaser.yaml around lines 38 to 41: the current owner/name fields use bare

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567191

- [review_comment] 2025-09-16T03:15:02Z by coderabbitai[bot] (.goreleaser.yaml:41)

```text
.goreleaser.yaml around lines 38 to 41: the current owner/name fields use bare
templates which fail YAML linting and are unnecessary for releasing to the same
repository; either remove the two lines entirely so GoReleaser infers the repo,
or if you must keep them wrap the template values in quotes (e.g. owner: "{{
.Env.GITHUB_REPOSITORY_OWNER }}" and name: "{{ .Env.GITHUB_REPOSITORY_NAME }}")
to produce valid YAML.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Wrapped templated owner/name values in quotes to satisfy YAML tooling. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Wrapped templated owner/name values in quotes to satisfy YAML tooling. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:494
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## .goreleaser.yaml around lines 43 to 45: there is an extra trailing blank line

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567194

- [review_comment] 2025-09-16T03:15:02Z by coderabbitai[bot] (.goreleaser.yaml:45)

```text
.goreleaser.yaml around lines 43 to 45: there is an extra trailing blank line
after the snapshot.name_template block; remove the blank line so the file ends
immediately after the "name_template: SNAPSHOT-{{ .ShortCommit }}" line to
satisfy yamllint and CI.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Removed stray blank line so file ends directly after the snapshot template. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Removed stray blank line so file ends directly after the snapshot template. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:530
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In .markdownlint.yaml at line 4 the config disables MD013 repo-wide; remove the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567196

- [review_comment] 2025-09-16T03:15:03Z by coderabbitai[bot] (.markdownlint.yaml:4)

```text
In .markdownlint.yaml at line 4 the config disables MD013 repo-wide; remove the
global "MD013: false" entry and instead apply targeted exceptions: delete that
line from the repo config and either add per-file overrides in the markdownlint
config for specific file globs, or keep MD013 enabled and use inline markers
(<!-- markdownlint-disable MD013 --> / <!-- markdownlint-enable MD013 -->),
per-line disables, or file front-matter to silence the rule only where long
lines are intentional.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Removed global MD013 disable and documented per-file opt-out approach. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Removed global MD013 disable and documented per-file opt-out approach. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:564
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## .vscode/extensions.json lines 1-6: The workspace recommendations only include

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567204

- [review_comment] 2025-09-16T03:15:03Z by coderabbitai[bot] (.vscode/extensions.json:6)

```text
.vscode/extensions.json lines 1-6: The workspace recommendations only include
the Go extension but this repo uses YAML and Docker files; update the
recommendations array to add "redhat.vscode-yaml" and
"ms-azuretools.vscode-docker" so VS Code suggests installing YAML and Docker
extensions. Keep existing entries, avoid duplicates, and leave
unwantedRecommendations untouched.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Added YAML and Docker extensions to VS Code workspace recommendations. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added YAML and Docker extensions to VS Code workspace recommendations. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:601
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In .vscode/settings.json around lines 9 to 13, the workspace is not enabling

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567208

- [review_comment] 2025-09-16T03:15:03Z by coderabbitai[bot] (.vscode/settings.json:13)

```text
In .vscode/settings.json around lines 9 to 13, the workspace is not enabling
staticcheck or key gopls analyses and also contains a trailing comma that can
break JSON; update the gopls settings to set "staticcheck": true and enable
analyses such as "nilness", "shadow", "unusedparams", and "unusedwrite" (as
appropriate for your project), and remove the trailing comma after the last
property so the JSON remains valid.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Enabled staticcheck plus key gopls analyses and tidied JSON. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Enabled staticcheck plus key gopls analyses and tidied JSON. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:637
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In CHANGELOG.md around lines 17 to 30, the current entry is a freeform

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567211

- [review_comment] 2025-09-16T03:15:03Z by coderabbitai[bot] (CHANGELOG.md:30)

```text
In CHANGELOG.md around lines 17 to 30, the current entry is a freeform
brain-dump and must be converted to "Keep a Changelog" style: split the content
into explicit sections such as Added, Changed, Fixed (and optionally
Removed/Deprecated) and move each bullet under the appropriate section, convert
informal bullets into concise changelog-style lines, and append PR references
(e.g. " (#123)") for each item — leave a placeholder for PR numbers to be filled
in once merged and add a short header with the release version and date; ensure
the final format matches other entries in the file and includes a
[request_verification] note to confirm PR numbers after merge.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Reworked Unreleased section to follow Keep a Changelog formatting with placeholders for PR refs. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Reworked Unreleased section to follow Keep a Changelog formatting with placeholders for PR refs. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:673
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In create_review_tasks.py around lines 22-24 (and also update the similar logic

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567213

- [review_comment] 2025-09-16T03:15:03Z by coderabbitai[bot] (create_review_tasks.py:24)

```text
In create_review_tasks.py around lines 22-24 (and also update the similar logic
at lines 27-35), the code lexically sorts task identifiers which misorders items
like "T10" vs "T9"; change the sort to extract the numeric portion and sort by
that numeric value instead. Implement a sort key that parses the integer from
each task string (e.g., regex or split to grab trailing digits), fallback to 0
if no number present, then use sorted(completed_tasks, key=that_numeric_key) and
slice [:12]; apply the same numeric-key sorting to the other block at lines
27-35.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Numeric sort helper ensures tasks order naturally (T9 before T10). |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Numeric sort helper ensures tasks order naturally (T9 before T10). Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:712
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In demos/responsive-tui.tape around lines 20 to 27 (and similarly at 81-88 and

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567216

- [review_comment] 2025-09-16T03:15:04Z by coderabbitai[bot] (demos/responsive-tui.tape:27)

```text
In demos/responsive-tui.tape around lines 20 to 27 (and similarly at 81-88 and
138-146), the script sets COLUMNS to simulate mobile layout but this is cosmetic
for most tools; update the demo to either display an explicit on-screen note
that setting COLUMNS is only cosmetic, or trigger the TUI's real layout switch
API/flag if available so the UI truly reflows; also reduce sleep/delay durations
used around these sections to speed CI. Ensure the note is visible before the
COLUMNS change (or replace the COLUMNS step with a proper layout switch), and
apply the same change to the other line ranges mentioned.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Demo now calls out COLUMNS overrides as cosmetic and trims long sleeps. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Demo now calls out COLUMNS overrides as cosmetic and trims long sleeps. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:750
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In dependency_analysis.py around lines 233 to 243, the infrastructure dict is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567217

- [review_comment] 2025-09-16T03:15:04Z by coderabbitai[bot] (dependency_analysis.py:243)

```text
In dependency_analysis.py around lines 233 to 243, the infrastructure dict is
missing many referenced components which causes the validator to fail and DAG
edges to be dropped; add missing infra stubs such as "plugin_runtime",
"event_sourcing", "idempotency_keys", "controller_runtime", "k8s_api" and any
other referenced components in the codebase (e.g., "service_mesh",
"policy_engine", "sidecar_injector", "ci_cd", "secrets_manager") as minimal
string descriptions or promote them to feature objects consistent with the
surrounding structure so the validator recognizes them and edges in the DAG
remain intact.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Added missing infrastructure stubs so dependency validation no longer drops edges. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added missing infrastructure stubs so dependency validation no longer drops edges. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:788
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deploy/docker-compose.yml around lines 19 to 22 the service volume mounts

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567222

- [review_comment] 2025-09-16T03:15:04Z by coderabbitai[bot] (deploy/docker-compose.yml:22)

```text
In deploy/docker-compose.yml around lines 19 to 22 the service volume mounts
reference ./data but the repository has deploy/deploy/data (deploy/data is
missing), so either move the directory deploy/deploy/data → deploy/data to match
the current ./data mount (preferred) or update the compose file mounts for
app-all and app-producer to point to the existing path (e.g.,
./deploy/data:/data or an absolute path). Ensure the chosen fix is applied
consistently for all services and update any related documentation or .gitignore
entries if paths change.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Verified compose now mounts `./deploy/data` after relocating the data directory. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Verified compose now mounts `./deploy/data` after relocating the data directory. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:827
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deploy/grafana/dashboards/work-queue.json around lines 6 to 9, the PromQL

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567230

- [review_comment] 2025-09-16T03:15:04Z by coderabbitai[bot] (deploy/grafana/dashboards/work-queue.json:9)

```text
In deploy/grafana/dashboards/work-queue.json around lines 6 to 9, the PromQL
uses incorrect aggregation syntax "sum(...) by (le)"; change it to use the "sum
by (le) (...)" form and wrap the rate() call inside that aggregation so the
histogram_quantile receives a properly aggregated timeseries (i.e., compute rate
on the bucket metric over 5m, then apply sum by (le) around that result before
calling histogram_quantile).
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Updated query now aggregates with `sum by (le)` outside the `rate()` call. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Updated query now aggregates with `sum by (le)` outside the `rate()` call. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:865
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deploy/grafana/dashboards/work-queue.json around lines 20-24 (and likewise

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567239

- [review_comment] 2025-09-16T03:15:04Z by coderabbitai[bot] (deploy/grafana/dashboards/work-queue.json:24)

```text
In deploy/grafana/dashboards/work-queue.json around lines 20-24 (and likewise
adjust panels at lines 31-35), the stat panels lack units, thresholds and
field/state mappings; update the "Circuit Breaker State" stat to include
field/value mappings (0 -> "Closed (OK)" with green, 1 -> "Open (Alert)" with
red, 2 -> "Half-Open (Warn)" with amber), add a unit (e.g., "state" or "none"),
and set a reduce/threshold/color scheme so the numerical values render as
labeled colored states; for the "active workers" stat set unit to "none", add
thresholds so >0 is green and =0 is red, and ensure both panels use appropriate
fieldConfig -> mappings and thresholds entries so the UI shows readable labels
and colors at a glance.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Circuit breaker and worker stat panels now include mappings, thresholds, and units. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Circuit breaker and worker stat panels now include mappings, thresholds, and units. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:901
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deploy/grafana/dashboards/work-queue.json around lines 26 to 29, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567245

- [review_comment] 2025-09-16T03:15:04Z by coderabbitai[bot] (deploy/grafana/dashboards/work-queue.json:29)

```text
In deploy/grafana/dashboards/work-queue.json around lines 26 to 29, the
Prometheus target uses a raw metric "queue_length" which emits one time series
per replica; change the query to aggregate across replicas per logical queue
(for example use sum by (queue)(queue_length) or sum without grouping then
group_by the queue label) so the panel shows one series per queue rather than
one per instance, and update the legend/labeling to include the queue label for
clarity.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Queue length panel now queries `sum by (queue) (queue_length)`, collapsing per-instance series. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Queue length panel now queries `sum by (queue) (queue_length)`, collapsing per-instance series. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:941
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## deployments/admin-api/k8s-redis.yaml lines 1-62: the manifest lacks any

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567261

- [review_comment] 2025-09-16T03:15:05Z by coderabbitai[bot] (deployments/admin-api/k8s-redis.yaml:62)

```text
deployments/admin-api/k8s-redis.yaml lines 1-62: the manifest lacks any
pod/container security hardening; add a podSecurityContext and container
securityContext plus a locked ServiceAccount: set
spec.template.spec.serviceAccountName to a restricted SA (create a dedicated
redis SA with minimal RBAC), add spec.template.spec.securityContext with
runAsNonRoot: true and fsGroup (e.g., 1000) so /data is writable, and in the
container securityContext set runAsUser (non-root UID),
allowPrivilegeEscalation: false, capabilities: drop: ["ALL"],
readOnlyRootFilesystem: false (since Redis needs /data) and seccompProfile: {
type: RuntimeDefault }; ensure the SA and any PSP/PSP-replacement (PodSecurity
admission) policies permit these constraints.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Manifest now provisions a dedicated ServiceAccount and applies pod/container security contexts. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Manifest now provisions a dedicated ServiceAccount and applies pod/container security contexts. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:978
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/admin-api/k8s-redis.yaml around lines 62 to 74, the Service

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567278

- [review_comment] 2025-09-16T03:15:05Z by coderabbitai[bot] (deployments/admin-api/k8s-redis.yaml:74)

```text
In deployments/admin-api/k8s-redis.yaml around lines 62 to 74, the Service
manifest is missing a trailing newline and references the namespace
"redis-work-queue" without creating it; add a newline at EOF to satisfy the
linter and either add a Namespace manifest (apiVersion: v1, kind: Namespace,
metadata.name: redis-work-queue) in this file or a separate manifest applied
with the Service, or alternatively add a short comment/README note stating that
"redis-work-queue" must be created beforehand.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Namespace resource added alongside Service, satisfying manifest lint. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Namespace resource added alongside Service, satisfying manifest lint. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:1021
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/01_product_roadmap.md around lines 34 to 39, the roadmap dates list RC

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567287

- [review_comment] 2025-09-16T03:15:05Z by coderabbitai[bot] (docs/01_product_roadmap.md:39)

```text
In docs/01_product_roadmap.md around lines 34 to 39, the roadmap dates list RC
and GA without mapping which CI checks/gates must pass; update the roadmap to
enumerate the required checks and link them to the Release Plan and CI jobs.
Specifically, add a short checklist under the RC/GA entries naming the exact CI
jobs that must be green (unit, integration, e2e-with-redis, security-scan,
performance-smoke, deploy-preview), reference the canonical release-plan
file/path and the CI workflow file(s) (e.g., .github/workflows/release.yml), and
note any gating rules (e.g., “all listed jobs must pass and branch protection
enabled”) so the roadmap dates align with the Release Plan and CI gates.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | RC/GA bullets now enumerate gating CI jobs and reference release artifacts. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> RC/GA bullets now enumerate gating CI jobs and reference release artifacts. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:1059
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/01_product_roadmap.md around lines 48–52, the "Dependencies" bullets are

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567299

- [review_comment] 2025-09-16T03:15:05Z by coderabbitai[bot] (docs/01_product_roadmap.md:52)

```text
In docs/01_product_roadmap.md around lines 48–52, the "Dependencies" bullets are
vague; update each dependency to name the responsible owner and the concrete
artifact(s) that satisfy it (e.g., PR number, spec/doc link, or task ID).
Replace the three bullets with explicit entries such as "Tracing propagation —
owner: @alice — depends on finalized Job struct (PR #123) and processor API doc
(docs/processor-api.md#version-2)", "Reaper improvements — owner: @bob — depends
on reliable heartbeat semantics (RFC #45 / task JIRA-678)", and "Performance
tuning — owner: @carol — depends on priority dequeue semantics (PR #130) and
metrics completeness (metrics/README or dashboards PR #140)"; keep formatting
consistent with the file and add links or references where available.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Dependencies list now names owners and the exact upstream artifacts. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Dependencies list now names owners and the exact upstream artifacts. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:1098
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/03_milestones.md around lines 6 to 8, the milestone entries lack

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567307

- [review_comment] 2025-09-16T03:15:05Z by coderabbitai[bot] (docs/03_milestones.md:8)

```text
In docs/03_milestones.md around lines 6 to 8, the milestone entries lack
assigned owners/DRIs; add an owner for each milestone (name, role, contact) and
a backup/secondary DRI, and include a one-line responsibility statement per
owner. Update the milestones list or table to add an "Owner / DRI" column or a
subheading under each milestone with the owner's name, role, email/Slack handle,
and their specific accountability (e.g., delivery lead, QA lead), and ensure
dependencies or decision gates note who is responsible for sign-off.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Every milestone now lists owner, backup DRI, and responsibility statement. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Every milestone now lists owner, backup DRI, and responsibility statement. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:28
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/04_sprint_plans.md around lines 5 to 8, replace the ambiguous term

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567313

- [review_comment] 2025-09-16T03:15:05Z by coderabbitai[bot] (docs/04_sprint_plans.md:8)

```text
In docs/04_sprint_plans.md around lines 5 to 8, replace the ambiguous term
"bi-weekly sprints" with the explicit phrase "two-week sprints" (and any other
occurrences in this file) so the plan unambiguously states sprint length; update
the sentence to read something like "Four two-week sprints lead to v1.0.0." and
verify surrounding text remains grammatically correct.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 10 | Terminology now states "two-week sprints" explicitly. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Terminology now states "two-week sprints" explicitly. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:65
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/06_technical_spec.md around lines 113–117, the reaper section is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567317

- [review_comment] 2025-09-16T03:15:06Z by coderabbitai[bot] (docs/06_technical_spec.md:117)

```text
In docs/06_technical_spec.md around lines 113–117, the reaper section is
hand-wavy: stop inferring original queue from payload and instead persist an
origin_queue field in job metadata; limit SCAN cost by using SCAN/SSCAN with
COUNT plus a per-invocation time budget and randomized jitter between scans to
avoid thundering-herd effects; and perform RPOP/LPUSH re-queue operations inside
a Redis Lua script (EVAL) so list mutation and heartbeat checks are atomic and
consistent. Ensure the spec describes how the reaper reads heartbeat keys, skips
live workers, uses the persisted origin_queue to determine destination list,
enforces a page/time limit per run, and sleeps with jitter between pages.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Spec now details origin queue persistence, bounded SCAN windows with jitter, and Lua-based requeue steps. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Spec now details origin queue persistence, bounded SCAN windows with jitter, and Lua-based requeue steps. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:100
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/13_release_versioning.md around lines 21 to 25, the release checklist

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567321

- [review_comment] 2025-09-16T03:15:06Z by coderabbitai[bot] (docs/13_release_versioning.md:25)

```text
In docs/13_release_versioning.md around lines 21 to 25, the release checklist
lacks supply-chain verification gates; add a new checklist item requiring an
SBOM, provenance (SLSA/OIDC) attestation, and signed artifacts (cosign) before
release. Update the numbered list to include a clear line such as “4) Ensure
supply-chain artifacts present: SBOM generated, build provenance/SLSA
attestations (e.g. OIDC) available, and release artifacts signed (e.g. cosign)”;
optionally add brief links or references to existing internal tooling or
standards used for SBOM/provenance/signing and ensure CI blocks release when
these artifacts are missing.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Release checklist now includes an explicit supply-chain verification step. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Release checklist now includes an explicit supply-chain verification step. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:139
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/13_release_versioning.md around lines 26 to 31, the current instructions

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567323

- [review_comment] 2025-09-16T03:15:06Z by coderabbitai[bot] (docs/13_release_versioning.md:31)

```text
In docs/13_release_versioning.md around lines 26 to 31, the current instructions
use a lightweight tag and recommend git push --tags which can push all local
tags; change to recommend creating an annotated or signed tag and pushing only
that single ref. Update the steps to show using git tag -a (or -s) with a
message, then git push origin <tag-name>, replacing the generic --tags flow so
only the new release tag is published.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Release guide now instructs creating annotated tags and pushing the single ref. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Release guide now instructs creating annotated tags and pushing the single ref. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:178
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/YOU ARE WORKER 6.rb around lines 3 to 4, the README uses vendor-specific

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567336

- [review_comment] 2025-09-16T03:15:06Z by coderabbitai[bot] (docs/YOU ARE WORKER 6.rb:4)

```text
In docs/YOU ARE WORKER 6.rb around lines 3 to 4, the README uses vendor-specific
phrasing ("Claude worker"); replace that with a neutral role description (e.g.,
"a worker in the SLAPS task execution system" or "task worker") so the text
reads project-neutral: update the sentence to remove the vendor name and ensure
it still communicates that the role claims and executes tasks for the
go-redis-work-queue project.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Role description now uses vendor-neutral wording and adds safety guidance. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Role description now uses vendor-neutral wording and adds safety guidance. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:251
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/YOU ARE WORKER 6.rb around lines 6 to 13, the workflow step that tells

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567349

- [review_comment] 2025-09-16T03:15:06Z by coderabbitai[bot] (docs/YOU ARE WORKER 6.rb:13)

```text
In docs/YOU ARE WORKER 6.rb around lines 6 to 13, the workflow step that tells
workers to "Claim a task by moving it to slaps-coordination/claude-001/" lacks
any caution about mv being non-atomic across filesystems and about race
conditions that can corrupt the queue; add a brief safety note instructing
maintainers to (1) prefer an atomic rename on the same filesystem or implement
an explicit locking/claim protocol (e.g., create a unique temporary/claim file
or use link+unlink or filesystem locks) to avoid collisions, (2) detect and
handle failures from mv/rename and retry/abort safely, and (3) fall back to an
explicit copy+atomic-rename pattern with checks when crossing filesystems to
prevent duplicate processing.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Added move safety guidance outlining atomic renames, copy+rename fallbacks, and error checks. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added move safety guidance outlining atomic renames, copy+rename fallbacks, and error checks. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:287
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/YOU ARE WORKER 6.rb around lines 21 to 26, the example shell commands

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350567354

- [review_comment] 2025-09-16T03:15:06Z by coderabbitai[bot] (docs/YOU ARE WORKER 6.rb:26)

```text
In docs/YOU ARE WORKER 6.rb around lines 21 to 26, the example shell commands
use paths with spaces unquoted and a plain mv that can clobber files; update the
examples to quote all filesystem paths (e.g.
"slaps-coordination/open-tasks/P1.T001.json") and add the -n flag to mv (mv -n
"source" "dest/") to prevent overwriting; ensure any path that could be
interpreted as an option is protected by quoting or by using -- where
appropriate.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Documentation now demonstrates safe quoted `mv -n` usage. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Documentation now demonstrates safe quoted `mv -n` usage. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:327
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In auto_commit.sh around line 1, the script lacks Bash strict mode which can

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569555

- [review_comment] 2025-09-16T03:16:13Z by coderabbitai[bot] (auto_commit.sh:1)

```text
In auto_commit.sh around line 1, the script lacks Bash strict mode which can
hide failures in the infinite git loop; add a strict-mode invocation immediately
after the shebang and set a safe IFS: enable errexit, errtrace, nounset and
pipefail (i.e. set -Eeuo pipefail) and set IFS to newline+tab so word-splitting
is safe; ensure this is the first executable code so the script exits on errors
and treats unset variables as failures.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Script now enables `set -Eeuo pipefail` and sets IFS safely right after the shebang. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Script now enables `set -Eeuo pipefail` and sets IFS safely right after the shebang. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:364
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In auto_commit.sh around lines 4-6 (and similarly lines 45-47) the script starts

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569560

- [review_comment] 2025-09-16T03:16:14Z by coderabbitai[bot] (auto_commit.sh:6)

```text
In auto_commit.sh around lines 4-6 (and similarly lines 45-47) the script starts
an unconditional infinite loop which can DOS CI; add a guard using an
environment variable (e.g., MAX_ITERATIONS) with a sensible default and an
iteration counter that increments each loop and breaks when the max is reached,
and add signal handling: install a trap for SIGINT and SIGTERM that sets a flag
(or exits) so the loop can terminate cleanly; ensure the loop checks the flag
before each iteration and after sleep, and prefer a configurable SLEEP_SECONDS
(default 300) so cadence remains adjustable.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Loop now honours MAX_ITERATIONS, sleep interval overrides, and signal handlers. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Loop now honours MAX_ITERATIONS, sleep interval overrides, and signal handlers. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:400
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In auto_commit.sh around lines 30 to 41, the script currently parses git push

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569563

- [review_comment] 2025-09-16T03:16:14Z by coderabbitai[bot] (auto_commit.sh:41)

```text
In auto_commit.sh around lines 30 to 41, the script currently parses git push
output with grep which is brittle; instead call git rev-parse --abbrev-ref HEAD
to get the current branch, run git push origin <current-branch> (or git push
--set-upstream origin <current-branch> if upstream is not set) and check git’s
exit status ($?) to determine success; remove the grep pipeline and use the push
command’s exit code to log success or failure, and when upstream is unset detect
that (e.g., by checking git rev-parse --symbolic-full-name @{u} or examining
push exit code) and set upstream only when needed.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Push logic now checks git exit codes directly and handles upstream detection without grepping output. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Push logic now checks git exit codes directly and handles upstream detection without grepping output. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:438
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In config/config.example.yaml around line 2, the default Redis address is set to

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569570

- [review_comment] 2025-09-16T03:16:14Z by coderabbitai[bot] (config/config.example.yaml:2)

```text
In config/config.example.yaml around line 2, the default Redis address is set to
"redis:6379" which breaks local development; change the default to
"localhost:6379" and add a one-line comment explaining that Docker Compose uses
the service name (e.g., redis:6379) and show an example override instruction
(mention using a compose-specific config or environment variable to set addr to
redis:6379 for containerized runs). Ensure the example indicates how to override
the value (compose config or env) and keep the default local-friendly value in
the example file.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Default now targets localhost with a comment noting Docker overrides. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Default now targets localhost with a comment noting Docker overrides. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:476
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In config/config.example.yaml around lines 50 to 66, the idempotency settings

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569573

- [review_comment] 2025-09-16T03:16:14Z by coderabbitai[bot] (config/config.example.yaml:66)

```text
In config/config.example.yaml around lines 50 to 66, the idempotency settings
lack precise semantics; update the example with comments that (1) define exactly
what constitutes the idempotency key (e.g., which headers and/or payload hash
are used and the format that fills {queue},{tenant},{key}), (2) describe
collision behavior (what happens when two requests produce the same key — e.g.,
first-wins, overwrite, or atomic check-and-set) and how retries interact with
stored results, (3) clarify TTL semantics versus cleanup_interval (TTL is how
long a record is valid; cleanup is periodic garbage collection and may not
immediately remove expired keys), and (4) note implications for storage options
(e.g., Redis hashes vs keys, compression effects) so users can configure
default_ttl, cleanup_interval, max_retries and retry_delay correctly.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Added inline comments covering key format, collision behaviour, and TTL vs cleanup semantics. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added inline comments covering key format, collision behaviour, and TTL vs cleanup semantics. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:514
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In config/config.example.yaml around lines 67 to 80, the outbox section is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569576

- [review_comment] 2025-09-16T03:16:14Z by coderabbitai[bot] (config/config.example.yaml:80)

```text
In config/config.example.yaml around lines 67 to 80, the outbox section is
disabled and only notes "Requires database setup" without any example DB
configuration or usage guidance; add a commented example DSN and required
schema/table names (e.g., outbox table name, columns used), an example polling
query or SQL snippet, and recommended config knobs (connection string, table,
schema, retention/purge query) or a link to the docs page that defines the
outbox setup so users can enable and configure it easily.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Outbox config now includes sample DSN, table name, and polling query hints. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Outbox config now includes sample DSN, table name, and polling query hints. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:555
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In create_review_tasks.py around lines 1-4 and also update lines 102-112,

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569580

- [review_comment] 2025-09-16T03:16:14Z by coderabbitai[bot] (create_review_tasks.py:4)

```text
In create_review_tasks.py around lines 1-4 and also update lines 102-112,
convert the script into a CLI tool by wrapping execution in if __name__ ==
'__main__' and adding argparse flags --limit (int) and --dir (path) to control
output quantity and directory; add an optional --timestamp flag (ISO8601 or
epoch) that, when provided, is used instead of datetime.now() so outputs are
deterministic for CI/tests; refactor functions that currently call
datetime.now() or use globals to accept a timestamp parameter (defaulting to
now) and ensure the main entry parses flags, injects the parsed timestamp into
those functions, and writes outputs to the specified dir with behavior unchanged
when flags are omitted.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Script now exposes argparse flags for limit, dirs, and timestamp with a guarded main entry. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Script now exposes argparse flags for limit, dirs, and timestamp with a guarded main entry. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:592
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In create_review_tasks.py around lines 9 to 11, the code calls

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569586

- [review_comment] 2025-09-16T03:16:14Z by coderabbitai[bot] (create_review_tasks.py:11)

```text
In create_review_tasks.py around lines 9 to 11, the code calls
os.listdir(completed_dir) which will crash if completed_dir does not exist; add
a guard to check for directory existence (os.path.isdir or os.path.exists)
before listing, and either create the directory (os.makedirs(completed_dir,
exist_ok=True)) or skip processing when it’s missing; then proceed to iterate
over files only when the directory exists to avoid the crash.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Duplicate filter now lowercases filenames before checking. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Duplicate filter now lowercases filenames before checking. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:958
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In create_review_tasks.py around lines 30-31, the code constructs timestamps

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569589

- [review_comment] 2025-09-16T03:16:14Z by coderabbitai[bot] (create_review_tasks.py:31)

```text
In create_review_tasks.py around lines 30-31, the code constructs timestamps
with datetime.now().isoformat()+"Z" which produces a naive local time mislabeled
as UTC; change this to use an aware UTC timestamp by calling
datetime.now(timezone.utc).isoformat() (and add/import timezone from datetime if
missing) so the produced ISO string reflects real UTC (or, if you must keep the
trailing "Z", use datetime.now(timezone.utc).isoformat().replace("+00:00","Z")).
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Timestamps now derive from `datetime.now(timezone.utc)` and render with `Z`. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Timestamps now derive from `datetime.now(timezone.utc)` and render with `Z`. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:668
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In demos/lipgloss-transformation.tape around lines 136-142 (and similarly for

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569595

- [review_comment] 2025-09-16T03:16:15Z by coderabbitai[bot] (demos/lipgloss-transformation.tape:142)

```text
In demos/lipgloss-transformation.tape around lines 136-142 (and similarly for
147-162) you are scripting "Type" commands that will fail if the referenced
binary doesn’t exist; add a clear guard or disclaimer: check for the binary file
(e.g., test -x ./bin/queue-tui) before running the Type sequence and skip or
print an explanatory echo if missing, or replace the sequence with a
commented/echoed mock note indicating this is a scripted demo and not executed;
ensure the demo prints a clear message when skipped so users know why the
interactive steps weren’t run.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Tape now runs `if [ -x ./bin/queue-tui ]` before invoking the binary. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Tape now runs `if [ -x ./bin/queue-tui ]` before invoking the binary. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:704
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In demos/lipgloss-transformation.tape around lines 271 to 276, the script calls

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569601

- [review_comment] 2025-09-16T03:16:15Z by coderabbitai[bot] (demos/lipgloss-transformation.tape:276)

```text
In demos/lipgloss-transformation.tape around lines 271 to 276, the script calls
figlet directly which will break on hosts without figlet; wrap the figlet
invocation in a guard that checks for the figlet binary (e.g., `command -v
figlet >/dev/null`) and, if missing, output a sensible fallback (plain "Redis
TUI" or a simple ASCII alternative) so the tape continues; apply the same
guarded pattern used elsewhere in the repo for the other figlet invocation.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | figlet call now wrapped in `command -v` test with plain-text fallback. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> figlet call now wrapped in `command -v` test with plain-text fallback. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:742
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In demos/responsive-tui.tape around lines 271-278 the final figlet call can

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569614

- [review_comment] 2025-09-16T03:16:15Z by coderabbitai[bot] (demos/responsive-tui.tape:282)

```text
In demos/responsive-tui.tape around lines 271-278 the final figlet call can
throw when figlet isn’t installed; wrap the figlet invocation in a safe guard so
the demo won’t crash on clean systems by detecting availability (e.g., try/catch
around require/spawn or check for the binary) and providing a graceful fallback
(render plain text or a simpler ASCII header) when figlet isn’t present,
ensuring the tape continues without error.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In dependency_analysis.py around lines 7 to 231, feature keys use kebab-case

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569619

- [review_comment] 2025-09-16T03:16:15Z by coderabbitai[bot] (dependency_analysis.py:231)

```text
In dependency_analysis.py around lines 7 to 231, feature keys use kebab-case
(e.g., "admin-api") while dependency lists use snake_case (e.g., "admin_api"),
causing resolution failures; normalize names to one canonical form (pick either
kebab-case or snake_case) at import by mapping all feature keys and all
dependency entries through the same normalizer (replace - with _ or vice versa)
before building the graph, correct known typos (e.g., change
"distributed_tracing" to "distributed-tracing-integration" or its normalized
equivalent), and add a validation pass that checks each dependency resolves to a
defined feature and logs or raises on unresolved refs so downstream DAGs fail
fast.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Added normalization helpers, alias mapping, infrastructure stubs, and a validation routine. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added normalization helpers, alias mapping, infrastructure stubs, and a validation routine. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:814
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/05_architecture.md around line 7, the architecture text omits the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569626

- [review_comment] 2025-09-16T03:16:16Z by coderabbitai[bot] (docs/05_architecture.md:7)

```text
In docs/05_architecture.md around line 7, the architecture text omits the
exactly-once/idempotency component; update the paragraph to add an
“Exactly-once” component and describe its flows: record/check idempotency key at
job intake, guard worker processing with idempotency lookup, persist
side-effects to an outbox before acknowledging completion, and publish outbox
entries to external systems; also mention how the reaper and circuit breaker
interact with idempotency (do not double-process keys) and that observability
should include metrics/traces for idempotency/outbox operations.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Component list and data flows now call out the idempotency/outbox layer. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Component list and data flows now call out the idempotency/outbox layer. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:854
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/06_technical_spec.md around lines 124-129 and also 131-134, the metrics

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569642

- [review_comment] 2025-09-16T03:16:16Z by coderabbitai[bot] (docs/06_technical_spec.md:129)

```text
In docs/06_technical_spec.md around lines 124-129 and also 131-134, the metrics
and logging section lacks explicit label names, cardinality bounds, unit
verification, and stable log key rules; update the doc to (1) list every metric
label schema (e.g., queue_length{queue}) and state a max cardinality or allowed
value set for each label, (2) verify and declare that the histogram metric uses
seconds (or rename suffix) so `_seconds` matches actual units, (3) state exact,
enforced log key names (trace_id, span_id, job_id, queue, worker_id) and forbid
logging secrets or PII, and (4) add a short note about how to enforce these
constraints in code/review (e.g., validation rules or linter/checklist).
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Spec now enumerates metric labels, units, and canonical log keys with guidance. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Spec now enumerates metric labels, units, and canonical log keys with guidance. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:892
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/09_requirements.md around lines 43-49, the acceptance criteria are

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569663

- [review_comment] 2025-09-16T03:16:16Z by coderabbitai[bot] (docs/09_requirements.md:49)

```text
In docs/09_requirements.md around lines 43-49, the acceptance criteria are
currently vague; update them to (1) enumerate the exact metric names and
expected types exposed at /metrics (e.g., request_count: counter,
request_duration_seconds: histogram, job_queue_length: gauge,
worker_registered_total: gauge) so tests can verify them, (2) specify that
/readyz must return healthy only if a successful Redis PING is received and at
least one worker is registered (describe the exact probe: call Redis PING and
check worker registration API/state), and (3) require admin destructive commands
to prompt for confirmation interactively and also accept a --yes flag for
non-interactive runs; additionally require automated unit/integration tests that
assert metric names/types, the /readyz behavior under Redis failure and
no-worker conditions, and admin command behavior with and without --yes.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Acceptance criteria now spell out metric names, readiness semantics, and admin safeguards. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Acceptance criteria now spell out metric names, readiness semantics, and admin safeguards. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:931
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/12_performance_baseline.md around lines 26 to 28, the example starts a

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569678

- [review_comment] 2025-09-16T03:16:17Z by coderabbitai[bot] (docs/12_performance_baseline.md:28)

```text
In docs/12_performance_baseline.md around lines 26 to 28, the example starts a
Redis container without pinning a specific minor version and lacks teardown
steps; update the run command to use a fixed minor Redis tag (e.g.,
7.2.x-alpine) and add explicit cleanup commands (stop/remove or force remove) so
reviewers can reproduce results reliably and avoid leftover containers.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Baseline now pins Redis to 7.2.4-alpine and documents cleanup. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Baseline now pins Redis to 7.2.4-alpine and documents cleanup. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:973
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/12_performance_baseline.md around lines 31 to 33, the README tells users

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569692

- [review_comment] 2025-09-16T03:16:17Z by coderabbitai[bot] (docs/12_performance_baseline.md:33)

```text
In docs/12_performance_baseline.md around lines 31 to 33, the README tells users
to run Redis locally but the example/default config uses redis:6379 and
therefore mismatches; update the doc to explicitly instruct readers to set
redis.addr="localhost:6379" for this baseline (or update the example config to
use localhost:6379) so the instructions and config agree—prefer adding a
one-line note beneath the config block that says: "Note: set
redis.addr=\"localhost:6379\" if running Redis locally for this baseline."
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Documentation now highlights the localhost address needed for the baseline setup. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Documentation now highlights the localhost address needed for the baseline setup. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:1008
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/14_ops_runbook.md around lines 21 to 26, replace the single unpinned

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569713

- [review_comment] 2025-09-16T03:16:18Z by coderabbitai[bot] (docs/14_ops_runbook.md:26)

```text
In docs/14_ops_runbook.md around lines 21 to 26, replace the single unpinned
"docker build -t job-queue-system:local ." entry with a pinned, reproducible
multi-arch build and an alternative compose build flow: update the docs to show
a buildx command that specifies platforms (e.g., linux/amd64,linux/arm64), pins
base image via build args or explicit tags, enables inline cache
(BUILDKIT_INLINE_CACHE=1), and uses --pull (and --push if publishing) so images
are reproducible across architectures; also add a separate example showing how
to build the same image via docker compose build for parity with later compose
notes.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Ops runbook now documents buildx multi-arch builds and compose parity commands. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Ops runbook now documents buildx multi-arch builds and compose parity commands. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:1045
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/14_ops_runbook.md around lines 32 to 35, the env-var mapping description

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569730

- [review_comment] 2025-09-16T03:16:18Z by coderabbitai[bot] (docs/14_ops_runbook.md:35)

```text
In docs/14_ops_runbook.md around lines 32 to 35, the env-var mapping description
is too vague; replace the hand-wavy “upper snake case replaces dots” with
explicit examples and parsing notes: add 1–2 concrete mappings (e.g.,
WORKER_COUNT -> worker.count and REDIS_ADDR -> redis.addr), show the
transformation rule (dots -> underscores, keys uppercased), and note how
booleans (true/false/1/0) and durations/times are parsed (e.g., "30s" ->
duration) and any required quoting; update the list to include these exact
mappings and parsing expectations so operators know how to set env vars
unambiguously.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Added explicit mapping examples and parsing notes for environment overrides. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added explicit mapping examples and parsing notes for environment overrides. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:1084
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/14_ops_runbook.md around lines 38 to 42, the guidance currently

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569756

- [review_comment] 2025-09-16T03:16:19Z by coderabbitai[bot] (docs/14_ops_runbook.md:42)

```text
In docs/14_ops_runbook.md around lines 38 to 42, the guidance currently
documents /healthz, /readyz, /metrics but doesn't instruct how to restrict
access; update this section to recommend binding health/metrics endpoints to
localhost or a dedicated admin interface, or expose them on a separate
port/interface, and add explicit protection guidance: enforce network
policies/firewall rules to restrict access, require authentication/authorization
(mTLS, bearer tokens or HTTP basic+IP allowlist) for admin/metrics endpoints,
and note Prometheus should scrape via a securely proxied or authenticated
endpoint; keep the examples concise and state to avoid exposing these endpoints
on public listeners.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Ops runbook now calls out binding metrics/health endpoints to controlled interfaces and restricting access. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Ops runbook now calls out binding metrics/health endpoints to controlled interfaces and restricting access. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:1123
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/14_ops_runbook.md around lines 51 to 75, the purge-dlq example and

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569759

- [review_comment] 2025-09-16T03:16:19Z by coderabbitai[bot] (docs/14_ops_runbook.md:75)

```text
In docs/14_ops_runbook.md around lines 51 to 75, the purge-dlq example and
surrounding admin CLI docs lack a dry-run example and an explicit RBAC note;
update the purge-dlq command example to include a --dry-run (and keep --yes
separate) showing safe preview usage, and add a short sentence noting that purge
operations require admin RBAC (e.g., only users/roles with purge/delete
permissions may execute) and recommend running dry-run first before --yes; keep
the other examples unchanged.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Runbook now documents `--dry-run` usage and mentions RBAC requirements for purge operations. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Runbook now documents `--dry-run` usage and mentions RBAC requirements for purge operations. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:28
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/dlq-remediation-pipeline.md around lines 121 to 137, the sample

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569774

- [review_comment] 2025-09-16T03:16:19Z by coderabbitai[bot] (docs/api/dlq-remediation-pipeline.md:137)

```text
In docs/api/dlq-remediation-pipeline.md around lines 121 to 137, the sample
response mixes integer and float types and embeds unit labels only in keys (ms),
which causes churn for clients; standardize metric types and lock units by (1)
making counts strictly integers (jobs_processed, jobs_matched, actions_executed,
actions_successful, actions_failed, rate_limit_hits, circuit_breaker_trips), (2)
expressing timing metrics as numbers in milliseconds as integers
(classification_time_ms, action_time_ms, end_to_end_time_ms) or explicitly state
they are floats if sub-millisecond precision is required, (3) keeping ratios/hit
rates as floats between 0 and 1 (cache_hit_rate), (4) update the JSON example to
use consistent types and values accordingly, and (5) add a short typed schema
section immediately after the example listing each field name, its JSON type,
and the unit (e.g., "classification_time_ms: integer (ms)") so clients have an
authoritative contract.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Example response now uses integer millisecond values with a schema table. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Example response now uses integer millisecond values with a schema table. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:65
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/dlq-remediation-pipeline.md around lines 197 to 252, the matcher

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569775

- [review_comment] 2025-09-16T03:16:20Z by coderabbitai[bot] (docs/api/dlq-remediation-pipeline.md:266)

```text
In docs/api/dlq-remediation-pipeline.md around lines 197 to 252, the matcher
block uses free-form strings (e.g., "retry_count": "< 3", "job_type":
"business_hours") without a formal grammar or schema; add a clear BNF or JSON
Schema for matcher fields, enumerating allowed keys/types (error_pattern as
regex, job_type enum or pattern, retry_count as structured comparator object
with operator and integer, time windows as structured objects like {start, end}
or named set references), update the example to use the structured form, and
document validation/error responses (HTTP 4xx with specific field and error
messages) so callers can validate and avoid undefined behavior.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/dlq-remediation-pipeline.md around lines 541–606, the audit example

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569781

- [review_comment] 2025-09-16T03:16:20Z by coderabbitai[bot] (docs/api/dlq-remediation-pipeline.md:606)

```text
In docs/api/dlq-remediation-pipeline.md around lines 541–606, the audit example
exposes full payloads and redaction behavior is not documented; update the text
and example to state that audit payloads are redacted by default, configurable
via an audit_redaction setting, and show a redacted response example. Specify a
minimal canonical list of always-masked fields (e.g., ssn, email, phone,
full_name, address, credit_card, payment_token, auth_token, password) and note
that nested payload keys matching patterns are masked; replace the
before_state/after_state content in the JSON example with redacted placeholders
(e.g., "<redacted>") and add a short note pointing to the config section that
explains how to change redaction level and add custom fields.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Audit example now shows redacted payloads with guidance on configurable masks. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Audit example now shows redacted payloads with guidance on configurable masks. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:147
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/dlq-remediation-pipeline.md around lines 860 to 876, the WebSocket

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569789

- [review_comment] 2025-09-16T03:16:20Z by coderabbitai[bot] (docs/api/dlq-remediation-pipeline.md:876)

```text
In docs/api/dlq-remediation-pipeline.md around lines 860 to 876, the WebSocket
events section lacks details about authentication and backpressure/heartbeat
handling; update the doc to require a bearer/token query param or Authorization
header for the /ws/dlq-remediation/events endpoint and show the token format and
renewal behavior, specify heartbeat/ping semantics (client must respond to
server pings and send an application-level heartbeat every N seconds, include
ping/pong timeouts and reconnect guidance), and define a
slow-consumer/backpressure policy (per-connection send buffer limits,
server-side queue thresholds, and the chosen strategy: drop oldest messages vs.
close connection with a close code and reason), plus recommend monitoring
metrics and recommended defaults (buffer size, ping interval, timeout) so
implementers can prevent memory blowup.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | WebSocket documentation now covers auth, ping/pong expectations, slow-consumer limits, and metrics. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> WebSocket documentation now covers auth, ping/pong expectations, slow-consumer limits, and metrics. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:187
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/dlq-remediation-ui.md around line 9, the doc currently states the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569797

- [review_comment] 2025-09-16T03:16:20Z by coderabbitai[bot] (docs/api/dlq-remediation-ui.md:9)

```text
In docs/api/dlq-remediation-ui.md around line 9, the doc currently states the
API has no auth; change the implementation notes to require authentication, CSRF
for browser clients, and RBAC with default-deny: add a mandatory auth middleware
(JWT/OAuth session) on all DLQ remediation endpoints, enforce CSRF validation on
state-changing requests originating from browsers, implement role checks (e.g.,
require "dlq_admin" or specific capability to purge/modify DLQs) and return 403
by default for unauthorized users, and document the required roles, token scope,
and recommended audit logging for all purge operations.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Authentication section now mandates RBAC, CSRF, and bearer/session tokens. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Authentication section now mandates RBAC, CSRF, and bearer/session tokens. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:229
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/dlq-remediation-ui.md around lines 231 to 241, the API currently

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569804

- [review_comment] 2025-09-16T03:16:20Z by coderabbitai[bot] (docs/api/dlq-remediation-ui.md:241)

```text
In docs/api/dlq-remediation-ui.md around lines 231 to 241, the API currently
uses a query param confirm=true which is not secure; replace this with a signed
confirmation token mechanism, change the endpoint to require a JSON POST body
that contains an explicit boolean dry_run flag (must be provided) and the signed
confirmation token, and add a mandatory change_reason string field that will be
validated and persisted to logs; update request validation to reject
query-string confirmation, validate and verify the token signature/expiry,
enforce dry_run presence before executing real changes, and ensure the reason is
recorded in audit logs for every remediation action.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Purge-all now takes a POST body with `confirm`, structured filters, and requires an `Idempotency-Key` header. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Purge-all now takes a POST body with `confirm`, structured filters, and requires an `Idempotency-Key` header. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:188
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/dlq-remediation-ui.md around lines 387 to 394, the documented rate

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569811

- [review_comment] 2025-09-16T03:16:20Z by coderabbitai[bot] (docs/api/dlq-remediation-ui.md:394)

```text
In docs/api/dlq-remediation-ui.md around lines 387 to 394, the documented rate
limits are arbitrary and not tied to roles/endpoints; update the section to
specify rate limits per endpoint and per role, differentiating read (list/get)
vs destructive (purge/requeue) operations. Change the list to explicitly state
limits per endpoint group (e.g., list, bulk, individual) and add stricter,
per-principal limits for destructive endpoints (purge/requeue) with
rate/permission mappings (e.g., admin/service account vs regular user), and
document how limits are enforced (per-IP, per-token) and any burst/penalty
behavior.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Rate limit section now lists per-role/per-endpoint quotas and enforcement details. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Rate limit section now lists per-role/per-endpoint quotas and enforcement details. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:306
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/SLAPS/FINAL-POSTMORTEM.md around lines 8-9 (and also lines 324-336), the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569820

- [review_comment] 2025-09-16T03:16:20Z by coderabbitai[bot] (docs/SLAPS/FINAL-POSTMORTEM.md:9)

```text
In docs/SLAPS/FINAL-POSTMORTEM.md around lines 8-9 (and also lines 324-336), the
header shows "Success Rate: 100% (74 completed, 14 remaining)" but the body
computes 74/88 = 84% — make these consistent by either (A) replacing the header
with a clear completion metric like "Completion: 84% (74/88 completed)" or (B)
keeping "Success Rate: 100%" and clarifying it refers to attempted tasks (e.g.,
"Success Rate: 100% (74 succeeded of 74 attempted); 14 tasks not yet
attempted"), then update the related occurrences at lines 324-336 to match the
chosen wording and ensure the numeric labels (completed/attempted/remaining) are
accurate.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Postmortem now reports “Completion: 84% (74/88 tasks finished)” with consistent follow-up bullets. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Postmortem now reports “Completion: 84% (74/88 tasks finished)” with consistent follow-up bullets. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:345
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In README.md around lines 3 to 6, the intro mentions features without a clear,

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569828

- [review_comment] 2025-09-16T03:16:21Z by coderabbitai[bot] (README.md:6)

```text
In README.md around lines 3 to 6, the intro mentions features without a clear,
maintainable reference; add a link to a living feature matrix (e.g., a
FEATURES.md or a table in the repo/docs site) indicating
supported/experimental/deprecated status. Update the intro to include a short
parenthetical or sentence like "See the feature matrix: <relative-link>" and
create or point to that living document in the repo, ensuring it is kept
up-to-date and clearly lists feature statuses and notes.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | README now links to `docs/features-ledger.md` for the live feature matrix. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> README now links to `docs/features-ledger.md` for the live feature matrix. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:384
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In README.md around lines 38-49 there is a mismatch between the documented Go

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569836

- [review_comment] 2025-09-16T03:16:21Z by coderabbitai[bot] (README.md:56)

```text
In README.md around lines 38-49 there is a mismatch between the documented Go
version (Go 1.25+), go.mod (go 1.24.0) and CI (go-version: 1.25.x); update
go.mod to "go 1.25" to match README and CI, run "go mod tidy" locally to refresh
module files, commit the updated go.mod and go.sum, and push so CI (still set to
1.25.x) can verify the build; alternatively, if you prefer 1.24, change README
and CI to 1.24.x and then run go mod tidy and commit.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## README.md lines 123-145: the "Purge all (test keys)" admin command is presented

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569846

- [review_comment] 2025-09-16T03:16:21Z by coderabbitai[bot] (README.md:145)

```text
README.md lines 123-145: the "Purge all (test keys)" admin command is presented
without sufficient gating; update the README to clearly label this command as
"DEV ONLY" (or similar), and document that it requires an explicit additional
safety flag (e.g., --dev or --force-dev) and the --yes confirmation to run;
include a prominent one-line warning explaining it is destructive and should not
be used in production and show the exact safe invocation (e.g., with both --yes
and the extra dev-only flag) so readers know how to run it deliberately.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In README.md around lines 149 to 155, the docs claim metrics/health are served

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569852

- [review_comment] 2025-09-16T03:16:21Z by coderabbitai[bot] (README.md:155)

```text
In README.md around lines 149 to 155, the docs claim metrics/health are served
on port 9090 which conflicts with Prometheus' default; update the README to
change the default metrics/health port to a non-conflicting port (e.g., 9091 or
2112) and clearly document the potential clash with local Prometheus (include a
note explaining how to override the port or how to avoid collision), ensuring
both the endpoint URLs and any startup/config examples reflect the new default.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In README.md around lines 166 to 185, the provided docker run example starts the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350569857

- [review_comment] 2025-09-16T03:16:21Z by coderabbitai[bot] (README.md:185)

```text
In README.md around lines 166 to 185, the provided docker run example starts the
image with --role=all but does not mount the required configuration or provide
environment overrides, which will cause startup failures; update the example to
show a bind‑mount of the host config (e.g., -v
./config/config.yaml:/app/config/config.yaml) and/or pass an env file or
explicit ENV mappings (e.g., --env-file env.list or -e KEY=VALUE) so the
container has required config values, and add a short note telling users to
ensure their local config and env file exist before running.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In append_metadata.py around line 11, the script uses a hardcoded absolute path

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350571937

- [review_comment] 2025-09-16T03:17:27Z by coderabbitai[bot] (append_metadata.py:11)

```text
In append_metadata.py around line 11, the script uses a hardcoded absolute path
(/Users/james/...), which breaks on other machines and CI; replace it with a
configurable value: accept an --ideas-dir CLI argument (via argparse) and/or
read an IDEAS_DIR environment variable, falling back to a sensible relative
default (e.g., ./docs/ideas) and expanding user (~) with os.path.expanduser;
validate the path exists and fail with a clear error if not.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Script now takes `--ideas-dir` (env override) instead of a hardcoded path. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Script now takes `--ideas-dir` (env override) instead of a hardcoded path. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:568
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In auto_commit.sh around lines 8 to 10, the current use of "ls ... | wc -l | tr

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350571968

- [review_comment] 2025-09-16T03:17:27Z by coderabbitai[bot] (auto_commit.sh:10)

```text
In auto_commit.sh around lines 8 to 10, the current use of "ls ... | wc -l | tr
-d ' '" is brittle for filenames and empty directories; replace each count with
a robust find invocation that restricts depth and type. For example, use find
slaps-coordination/open-tasks -mindepth 1 -maxdepth 1 -type f | wc -l (and
similarly for finished-tasks and help-me) to accurately count files regardless
of names or locales; ensure any non-existent directory errors are suppressed
(e.g., redirect find's stderr to /dev/null) if needed.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Replaced ls/wc pipeline with a `find`-based helper that tolerates missing dirs. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Replaced ls/wc pipeline with a `find`-based helper that tolerates missing dirs. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:604
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In auto_commit.sh around lines 16–28, the git commit call currently silences

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350571987

- [review_comment] 2025-09-16T03:17:27Z by coderabbitai[bot] (auto_commit.sh:28)

```text
In auto_commit.sh around lines 16–28, the git commit call currently silences
errors with 2>/dev/null and embeds a third‑party marketing line plus a
Co-Authored-By footer; remove the stderr redirection so failures surface, and
remove the marketing/Co‑Authored‑By text from the default commit message (or
make those footers opt‑in via an environment variable like INCLUDE_GENERATOR),
then ensure the script checks git commit's exit code and exits non‑zero (or
prints the git error) on failure so CI/users see the problem.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Commit step now surfaces errors and uses a neutral message without co-author footers. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Commit step now surfaces errors and uses a neutral message without co-author footers. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:641
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In BUGS.md around lines 27-28: the repo currently depends on both

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572028

- [review_comment] 2025-09-16T03:17:28Z by coderabbitai[bot] (BUGS.md:28)

```text
In BUGS.md around lines 27-28: the repo currently depends on both
github.com/go-redis/redis/v8 and redis/go-redis/v9; pick v9 only, remove v8
references, and wrap the v9 client behind a small local interface (e.g. Cmdable)
used throughout the codebase for easier testing. Update all imports to
redis/go-redis/v9, replace any v8-specific types/usages with their v9
equivalents, add/adjust the interface file and constructor to return the
interface, update go.mod to remove the v8 dependency and run go mod tidy, and
fix tests/mocks to use the new interface so the duplicate dependency tree is
eliminated.
```

> [!INFO]- **Rejected**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | No | Yes | 5 | Migrating the entire codebase to redis/go-redis v9 is a large refactor tracked separately. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Reviewer note deemed non-actionable. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:678
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In claude_worker.py around line 35, the return type hint uses Optional[Path] but

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572046

- [review_comment] 2025-09-16T03:17:28Z by coderabbitai[bot] (claude_worker.py:55)

```text
In claude_worker.py around line 35, the return type hint uses Optional[Path] but
Optional is not imported; add the missing import to the top-level imports (e.g.,
from typing import Optional) so the type annotation resolves correctly and
static type checkers/runtime annotations won’t fail.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In claude_worker.py around lines 90-145 the worker currently blocks on

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572059

- [review_comment] 2025-09-16T03:17:28Z by coderabbitai[bot] (claude_worker.py:145)

```text
In claude_worker.py around lines 90-145 the worker currently blocks on
interactive input() calls to mark tasks done/help/failed, which turns the module
into a manual tool; replace the interactive loop with programmatic task status
handling: remove input() calls and instead call a task execution method (e.g.,
self.execute_task(task_data)) that returns a status enum/tuple (status,
summary_or_error), set timestamps/actor from that result, serialize and move the
file to finished/help/failed directories accordingly, and implement retries,
timeout handling, and logging for failures; alternatively, if the intention is
truly manual, rename the file/class to manual_task_tracker.py and update
docs/imports. Ensure no user prompts remain and keep file moves and JSON writes
identical to current behavior.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Worker now delegates to a pluggable executor returning TaskStatus instead of prompting via input(). |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Worker now delegates to a pluggable executor returning TaskStatus instead of prompting via input(). Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:755
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In claude_worker.py around lines 157-159 the bare "except: pass" silently

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572062

- [review_comment] 2025-09-16T03:17:28Z by coderabbitai[bot] (claude_worker.py:159)

```text
In claude_worker.py around lines 157-159 the bare "except: pass" silently
swallows all exceptions; replace it by catching and handling only the expected
exceptions (e.g., json.JSONDecodeError, OSError/IOError, ValueError) and log the
exception with context via the module logger, return False for
handled/non-critical errors, and re-raise truly critical exceptions
(KeyboardInterrupt, SystemExit, MemoryError) so they propagate; ensure logs
include the exception message and stack trace (logger.exception) to aid
debugging.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Replaced bare except with targeted exception handling and failure logging. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Replaced bare except with targeted exception handling and failure logging. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:796
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/admin-api/monitoring.yaml lines 1-66: this ConfigMap duplicates

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572067

- [review_comment] 2025-09-16T03:17:29Z by coderabbitai[bot] (deployments/admin-api/monitoring.yaml:66)

```text
In deployments/admin-api/monitoring.yaml lines 1-66: this ConfigMap duplicates
alerts already managed in deployments/kubernetes/monitoring.yaml and conflicts
with the monitoring operator (ConfigMap-based rules vs PrometheusRule). Fix by
either deleting this file entirely, or converting its contents into a
PrometheusRule (and ServiceMonitor if needed) CRD placed in the same namespace
and using the same labels/owner/namespace conventions as the existing monitoring
manifests under deployments/kubernetes/monitoring.yaml so the operator picks it
up; ensure you do not keep both ConfigMap and PrometheusRule definitions for the
same alerts to avoid duplication.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Converted alert ConfigMap into a PrometheusRule resource. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Converted alert ConfigMap into a PrometheusRule resource. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:834
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/admin-api/monitoring.yaml around line 5 (and also line 71), the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572071

- [review_comment] 2025-09-16T03:17:29Z by coderabbitai[bot] (deployments/admin-api/monitoring.yaml:5)

```text
In deployments/admin-api/monitoring.yaml around line 5 (and also line 71), the
namespace is set to "redis-work-queue" which conflicts with the expected
"work-queue"; update the namespace value(s) at those lines to the canonical
"work-queue" so all manifests/dashboards use the same namespace and avoid 404s.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Namespaces now match the `work-queue` convention in both the rule and dashboard manifests. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Namespaces now match the `work-queue` convention in both the rule and dashboard manifests. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:873
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/admin-api/monitoring.yaml around lines 58-65, the alert divides

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572072

- [review_comment] 2025-09-16T03:17:29Z by coderabbitai[bot] (deployments/admin-api/monitoring.yaml:65)

```text
In deployments/admin-api/monitoring.yaml around lines 58-65, the alert divides
by container_spec_memory_limit_bytes without matching labels which creates
cardinality/vector-matching issues; change the denominator to the
kube-state-metrics memory limit metric (e.g.
kube_pod_container_resource_limits_bytes{resource="memory"} or the equivalent
kube_pod_container_resource_limits{resource="memory", unit="byte"}) and perform
an explicit vector match so the usage and limit align (for example use
on(namespace,pod,container) or include identical pod/container selectors), or
remove the alert until you can implement correct label-matched limits.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Alert now joins usage with kube-state-metrics limits and guards against zero values. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Alert now joins usage with kube-state-metrics limits and guards against zero values. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:907
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/admin-api/monitoring.yaml around lines 82 to 99 (and also lines

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572075

- [review_comment] 2025-09-16T03:17:29Z by coderabbitai[bot] (deployments/admin-api/monitoring.yaml:99)

```text
In deployments/admin-api/monitoring.yaml around lines 82 to 99 (and also lines
~118-125), PromQL label matchers use single quotes (e.g. {job='admin-api'})
which is invalid; update every PromQL target in this file to use double quotes
for label values (e.g. {job="admin-api"}), including status regexes and any
other label matchers, and search/replace all occurrences across the file so all
targets use double-quoted label values.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | JSON dashboard targets now escape double quotes for all PromQL label matchers. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> JSON dashboard targets now escape double quotes for all PromQL label matchers. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:946
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/admin-api/monitoring.yaml around line 128, the file currently

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572077

- [review_comment] 2025-09-16T03:17:29Z by coderabbitai[bot] (deployments/admin-api/monitoring.yaml:128)

```text
In deployments/admin-api/monitoring.yaml around line 128, the file currently
lacks a trailing newline at EOF; open the file and add a single newline
character at the end (ensure the file ends with a single newline), then save so
the file ends with a proper newline.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Verified the manifest ends with a single newline to satisfy linters. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Verified the manifest ends with a single newline to satisfy linters. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:982
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/kubernetes/monitoring.yaml around lines 1 to 17, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572079

- [review_comment] 2025-09-16T03:17:29Z by coderabbitai[bot] (deployments/kubernetes/monitoring.yaml:17)

```text
In deployments/kubernetes/monitoring.yaml around lines 1 to 17, the
ServiceMonitor resource is using the wrong apiVersion; replace "apiVersion: v1"
with "apiVersion: monitoring.coreos.com/v1" so the ServiceMonitor CRD is
recognized, then validate the manifest (kubectl apply --dry-run=client or
kubectl apply) and ensure the Prometheus Operator CRDs are installed in the
cluster before applying.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | ServiceMonitor now declares `monitoring.coreos.com/v1` to match the CRD. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> ServiceMonitor now declares `monitoring.coreos.com/v1` to match the CRD. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:1016
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/kubernetes/monitoring.yaml around lines 72-84, the alert

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572084

- [review_comment] 2025-09-16T03:17:29Z by coderabbitai[bot] (deployments/kubernetes/monitoring.yaml:84)

```text
In deployments/kubernetes/monitoring.yaml around lines 72-84, the alert
expression divides container_memory_usage_bytes by
container_spec_memory_limit_bytes with mismatched labels and can divide by zero;
replace it to use the kube-state-metrics limits metric (e.g.
kube_pod_container_resource_limits_bytes or kube_pod_container_resource_limits)
and perform a proper vector match by namespace/pod/container (or use
on(namespace,pod,container) group_left if needed) and guard against zero limits
by filtering the limit metric to > 0 (or applying clamp_min(limit,1)) before
division so the resulting ratio is valid and safe for comparison.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Alert now divides usage by a namespace-scoped kube-state memory limit with clamp_min safeguards. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Alert now divides usage by a namespace-scoped kube-state memory limit with clamp_min safeguards. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:1052
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/kubernetes/monitoring.yaml around lines 96-107, the rule

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572088

- [review_comment] 2025-09-16T03:17:30Z by coderabbitai[bot] (deployments/kubernetes/monitoring.yaml:107)

```text
In deployments/kubernetes/monitoring.yaml around lines 96-107, the rule
currently uses a boolean expr so $value becomes 0/1 and the annotation shows
garbage; change the rule to (1) filter by namespace, (2) keep a boolean expr to
fire the alert:
(certmanager_certificate_expiration_timestamp_seconds{name="admin-api-tls",namespace="your-namespace"}
- time()) < 7*24*3600, and (3) update the annotation to display the
time-to-expiry by evaluating the time-left expression, e.g. use {{
humanizeDuration
(certmanager_certificate_expiration_timestamp_seconds{name="admin-api-tls",namespace="your-namespace"}
- time()) }} so the annotation shows remaining time instead of 0/1.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Alert now scopes to `namespace="work-queue"` and surfaces the remaining TTL via `humanizeDuration`. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Alert now scopes to `namespace="work-queue"` and surfaces the remaining TTL via `humanizeDuration`. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:1091
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/kubernetes/monitoring.yaml around lines 109 to 119, the alert

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572094

- [review_comment] 2025-09-16T03:17:30Z by coderabbitai[bot] (deployments/kubernetes/monitoring.yaml:119)

```text
In deployments/kubernetes/monitoring.yaml around lines 109 to 119, the alert
uses the metric http_requests_rate_limited_total which conflicts with other
files using rate_limit_exceeded_total; pick one canonical metric name (e.g.,
rate_limit_exceeded_total) and update this alert to use that name, then search
the repo for the other occurrences and replace them to the chosen name so all
rules/dashboards/queries are consistent; ensure the metric label set and any
recording rules or exporters that emit the metric are updated to match the new
canonical name and run a quick lint/test of Prometheus rules to verify no
references remain to the old name.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Alert and dashboard now reference the canonical `rate_limit_exceeded_total` metric. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Alert and dashboard now reference the canonical `rate_limit_exceeded_total` metric. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:1131
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/kubernetes/monitoring.yaml around line 220, the file is missing a

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572101

- [review_comment] 2025-09-16T03:17:30Z by coderabbitai[bot] (deployments/kubernetes/monitoring.yaml:220)

```text
In deployments/kubernetes/monitoring.yaml around line 220, the file is missing a
trailing newline at EOF; add a single newline character at the end of the file
so the final line terminator is present and the file ends with a newline.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 6 | Verified the manifest now ends with a single newline to satisfy linters. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Verified the manifest now ends with a single newline to satisfy linters. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:28
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/README.md around lines 39-44 (and also update occurrences at

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572107

- [review_comment] 2025-09-16T03:17:30Z by coderabbitai[bot] (deployments/README.md:44)

```text
In deployments/README.md around lines 39-44 (and also update occurrences at
165-170), the docs inconsistently reference /health, /healthz and /readyz;
choose the Kubernetes conventions and make them consistent: use /healthz for
liveness and /readyz for readiness across the entire document, update the curl
examples and any runbook references accordingly, and verify no other places
still reference /health (or swap meanings) so probes and examples match the
chosen endpoints.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | README and deployment probes now consistently use `/healthz` and `/readyz`. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> README and deployment probes now consistently use `/healthz` and `/readyz`. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:61
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/README.md around lines 80 to 88, the README claims a rate limit

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572110

- [review_comment] 2025-09-16T03:17:30Z by coderabbitai[bot] (deployments/README.md:88)

```text
In deployments/README.md around lines 80 to 88, the README claims a rate limit
of “100 rpm, burst 20” but no environment variables or flags are documented to
control those values; add explicit env config entries for the rate limiter
(e.g., RATE_LIMIT_RPM, RATE_LIMIT_BURST, optionally RATE_LIMIT_WINDOW_SECONDS
and RATE_LIMIT_ENABLED) with clear descriptions and sensible defaults (100, 20,
60, true), specify expected types (integer/boolean), and note that the
application should read these envs to configure the limiter; update the table to
include these variables, their descriptions and defaults so users can actually
tune the rate limiting behavior.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Rate limiter environment variables and defaults are now documented. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Rate limiter environment variables and defaults are now documented. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:98
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/README.md around lines 93 to 100, the current example uses

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572115

- [review_comment] 2025-09-16T03:17:30Z by coderabbitai[bot] (deployments/README.md:100)

```text
In deployments/README.md around lines 93 to 100, the current example uses
kubectl create secret ... --from-literal with real tokens which leaks secrets to
shell history and CI logs; update the docs to instruct creating secrets from
files or stdin (store tokens in files with restrictive permissions or pass via
pipe/stdin), or generate a YAML manifest with kubectl --dry-run=client -o yaml
and apply that manifest, and emphasise not to paste secrets directly into
shell/CI logs and to use environment variables or secret managers in CI.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | All admin API manifests and deployment defaults now point to the `work-queue` namespace. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> All admin API manifests and deployment defaults now point to the `work-queue` namespace. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:174
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/README.md around lines 95–100 the secret is created with -n

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572119

- [review_comment] 2025-09-16T03:17:30Z by coderabbitai[bot] (deployments/README.md:100)

```text
In deployments/README.md around lines 95–100 the secret is created with -n
work-queue but other admin-api manifests use -n redis-work-queue, causing
namespace drift; pick the cluster's canonical namespace (confirm with kubectl
get namespaces or team docs) and make it consistent across all affected files:
either change this README secret to -n redis-work-queue, or update every
manifest/script that uses redis-work-queue to work-queue. Specifically update
deployments/README.md; deployments/admin-api/* (k8s-deployment.yaml,
k8s-redis.yaml, monitoring.yaml, deploy.sh); deployments/kubernetes/*
(admin-api-deployment.yaml, rbac-token-service-deployment.yaml,
monitoring.yaml); deployments/scripts/* (setup-monitoring.sh,
test-staging-deployment.sh); and update docs/monitoring to reflect the chosen
namespace. Ensure all kubectl commands, YAML metadata.namespace fields, and Helm
values (if any) match the chosen namespace and run a dry validation (kubectl
apply --server-dry-run or kubeval) to confirm no mismatches remain.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | All admin API manifests and deployment defaults now point to the `work-queue` namespace. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> All admin API manifests and deployment defaults now point to the `work-queue` namespace. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:174
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/README.md around lines 146-148, the docs mention rate limits but

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572123

- [review_comment] 2025-09-16T03:17:31Z by coderabbitai[bot] (deployments/README.md:167)

```text
In deployments/README.md around lines 146-148, the docs mention rate limits but
omit that /metrics must not be internet-facing; update the notes to state
explicitly that the metrics endpoint must be exposed only via a ClusterIP-only
Service (no Ingress/LoadBalancer) and protected with a NetworkPolicy restricting
access to Prometheus scrape targets, and document that scraping should be done
via a ServiceMonitor or Prometheus scrape config targeting the ClusterIP service
only.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/00_assessment.md around lines 20–21, the doc currently pins "go-redis

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572129

- [review_comment] 2025-09-16T03:17:31Z by coderabbitai[bot] (docs/00_assessment.md:21)

```text
In docs/00_assessment.md around lines 20–21, the doc currently pins "go-redis
v8" — update wording to "go-redis v9 (v9+)" and add a concise migration
checklist that lists the breaking changes to account for (pipelines not
thread-safe; changed timeout/cancel behavior; Pipeline.Close/WithContext
removed; option renames MaxConnAge→ConnMaxLifetime and
IdleTimeout→ConnMaxIdleTime; connection reaper removal in favor of MaxIdleConns;
redis.Z type change from *Z→Z; reworked hooks API including DialHook; RESP3
behavior differences). Also append a short upgrade plan (steps: upgrade
dependency in lockfile, run tests, audit pipeline usages and option names,
update hooks and types, run performance/RESP3 smoke tests) plus a rollback
justification (pin v8 in lockfile and revert dependency if issues) or explicit
justification for staying on v8. Ensure the text stays brief, actionable, and
includes a checklist and upgrade/rollback plan.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Assessment now highlights go-redis v9 and documents migration/rollback steps. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Assessment now highlights go-redis v9 and documents migration/rollback steps. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:255
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/07_test_plan.md around lines 27 to 29, replace the vague "Chaos (where

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572137

- [review_comment] 2025-09-16T03:17:31Z by coderabbitai[bot] (docs/07_test_plan.md:29)

```text
In docs/07_test_plan.md around lines 27 to 29, replace the vague "Chaos (where
feasible in CI)" note with a deterministic list of fault-injection scenarios and
explicit pass/fail criteria: enumerate concrete failures (e.g., Redis SIGSTOP
for 30s, introduce 200ms p95 latency using tc netem on Redis port, inject 5% TCP
connection resets via iptables or tc loss), provide exact commands or CI steps
to run each injection (so they can be replayed in CI or locally), define how
long each injection should run and the sequence/timing, and state clear
pass/fail criteria for each (e.g., service stays healthy, no data loss, retries
succeed within X seconds, error rate < Y%) so reviewers and CI can
deterministically validate outcomes.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Chaos section now lists explicit failures, commands, durations, and pass/fail criteria. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Chaos section now lists explicit failures, commands, durations, and pass/fail criteria. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:298
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/07_test_plan.md around lines 41 to 45, the benchmark notes lack

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572146

- [review_comment] 2025-09-16T03:17:31Z by coderabbitai[bot] (docs/07_test_plan.md:45)

```text
In docs/07_test_plan.md around lines 41 to 45, the benchmark notes lack
reproducibility details; update the test plan to pin the GH runner size (exact
VM type/VM image), specify the exact Go version used, set and document
GOMAXPROCS (and recommend exporting it in the runner), fix and export the RNG
seed used by the synthetic job producer, document CPU governor/settings used
during runs, and ensure each run prints the commit SHA (and any build
tags/flags) and exports the seed so results can be reproduced; include these
fields in the Reporting section so CPU/memory/Redis/queue metrics are captured
alongside runner size, GOMAXPROCS, seed, CPU governor, and Go version.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Benchmark section now specifies runner type, Go version, GOMAXPROCS, seed, and reporting fields. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Benchmark section now specifies runner type, Go version, GOMAXPROCS, seed, and reporting fields. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:338
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/anomaly-radar-slo-budget.md around lines 124 to 176, the duration

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572177

- [review_comment] 2025-09-16T03:17:32Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:176)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 124 to 176, the duration
field "window" is ambiguous as currently shown ("720h0m0s"); update the docs to
explicitly state that durations are encoded as Go time.Duration strings and list
accepted formats (e.g., "72h", "720h0m0s", "30m", "1h30m", "1500ms"), include a
short note about parsing behavior (supports negative values and sub-second units
like "1500ms" or "1.5s"), and add one or two alternate example values in the
JSON response to demonstrate valid variants so downstream clients know how to
parse them.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Added explicit RFC3339/UTC notes under every response example. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added explicit RFC3339/UTC notes under every response example. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:808
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/anomaly-radar-slo-budget.md around lines 186 to 217, the config

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572224

- [review_comment] 2025-09-16T03:17:33Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:217)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 186 to 217, the config
endpoint documentation is missing required Authentication/Authorization details
for this admin surface; update the endpoint docs to include an
"Authentication/Authorization" block that states the endpoint requires admin
scope and specifies the auth scheme (Bearer JWT) consistent with the Admin API
docs, include required header (Authorization: Bearer <token>), required
roles/scopes (e.g., "admin" or specific scope name used by the Admin API), and a
short example note on denied responses (401/403) so readers know auth is
mandatory.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Endpoint doc now calls out Bearer auth requirements and 401/403 behaviours. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Endpoint doc now calls out Bearer auth requirements and 401/403 behaviours. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:415
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/canary-deployments.md around lines 15 to 19, the authentication

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572234

- [review_comment] 2025-09-16T03:17:33Z by coderabbitai[bot] (docs/api/canary-deployments.md:19)

```text
In docs/api/canary-deployments.md around lines 15 to 19, the authentication
headers listed are inconsistent with the rest of the repo (Bearer + X-API-Key).
Update the section to standardize on Authorization: Bearer <token> as the
primary method, replace X-API-Key with the repository-preferred alias
X-API-Token (if an alias is needed), and add one sentence clarifying semantics
and precedence (Authorization header takes precedence over X-API-Token when both
are present). Ensure the examples and header names in this block are changed
accordingly.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Auth section now aligns with repo conventions (Bearer primary, `X-API-Token` alias). |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Auth section now aligns with repo conventions (Bearer primary, `X-API-Token` alias). Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:454
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/canary-deployments.md around lines 303–345, the JSON mixes units

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572245

- [review_comment] 2025-09-16T03:17:33Z by coderabbitai[bot] (docs/api/canary-deployments.md:345)

```text
In docs/api/canary-deployments.md around lines 303–345, the JSON mixes units
(error_rate shown as 0.96 while success_rate is 99.04) — pick a single
convention and make fields explicit; update the schema to use percentage fields
(rename error_rate → error_percent and success_rate → success_percent) and set
their values to (error_count/job_count)*100 and (success_count/job_count)*100
respectively (rounded consistently, e.g., two decimal places), and add a short
comment/inline note in the docs that these fields are percentages to avoid
future confusion.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Metrics examples now expose `error_percent`/`success_percent` and clarify 0–100 scaling. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Metrics examples now expose `error_percent`/`success_percent` and clarify 0–100 scaling. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:493
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/canary-deployments.md around lines 556 to 592, the Deployment Object

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572255

- [review_comment] 2025-09-16T03:17:34Z by coderabbitai[bot] (docs/api/canary-deployments.md:592)

```text
In docs/api/canary-deployments.md around lines 556 to 592, the Deployment Object
"id" field is underspecified; update the schema and prose to explicitly state
the expected ID format (e.g., "uuid" v4 or "ULID"), allowed characters and
length, give a concrete example, and indicate that the server validates IDs and
will return HTTP 400 on invalid IDs; add a short note on the validation rules
(regex or exact format), whether IDs are case-sensitive, and link to the API
error response shape for 400 validation errors.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Deployment object now shows a ULID example and documents validation regex + 400 behaviour. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Deployment object now shows a ULID example and documents validation regex + 400 behaviour. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:532
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/capacity-planning-api.md around lines 146 to 158 the example calls

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572261

- [review_comment] 2025-09-16T03:17:34Z by coderabbitai[bot] (docs/api/capacity-planning-api.md:169)

```text
In docs/api/capacity-planning-api.md around lines 146 to 158 the example calls
calc.Calculate(..., metrics) but never declares metrics, causing copy-paste
compile errors; either add a minimal declaration such as metrics :=
capacityplanning.Metrics{ /* fill required fields */ } immediately before the
call, or remove the metrics parameter from the example call and adjust the
argument list accordingly so the snippet compiles.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/capacity-planning-api.md around lines 311 to 318, the import uses a

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572269

- [review_comment] 2025-09-16T03:17:34Z by coderabbitai[bot] (docs/api/capacity-planning-api.md:318)

```text
In docs/api/capacity-planning-api.md around lines 311 to 318, the import uses a
hardcoded placeholder module path "github.com/yourorg/..." which will mislead
users; update the import to either the repository's actual Go module path or
replace it with a neutral placeholder comment (e.g., // replace with your module
path) and show an example like module/path/to/automatic-capacity-planning so
readers know to substitute their own module path; ensure the docs clearly state
to replace the placeholder with the user's real module path.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Import example now references the canonical repo path with guidance to swap in local module names. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Import example now references the canonical repo path with guidance to swap in local module names. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:607
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/exactly-once-admin.md around lines 299 to 321, replace the literal

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572287

- [review_comment] 2025-09-16T03:17:34Z by coderabbitai[bot] (docs/api/exactly-once-admin.md:321)

```text
In docs/api/exactly-once-admin.md around lines 299 to 321, replace the literal
"YOUR_TOKEN" used in the curl examples with a clearly identifiable placeholder
(e.g. <YOUR_API_TOKEN_HERE> or ${API_TOKEN}) and add a brief note in the
examples that this must be substituted with a real token or read from an
environment variable; update all four curl snippets to use the chosen
placeholder consistently so readers cannot copy-paste a literal string by
mistake.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Curl examples now require `${API_TOKEN}` placeholders with guidance to source real tokens securely. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Curl examples now require `${API_TOKEN}` placeholders with guidance to source real tokens securely. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:645
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/PRD.md around lines 162-168 the current metric definition

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350572294

- [review_comment] 2025-09-16T03:17:34Z by coderabbitai[bot] (docs/PRD.md:168)

```text
In docs/PRD.md around lines 162-168 the current metric definition
queue_length{queue=...} risks unbounded label cardinality; update the spec to
require limiting/sanitizing queue labels by adding a new config entry (e.g.
metrics.allowed_queues: list of permitted queue names) and a fallback strategy
(hash or bucket remaining queues into a small fixed set or use a generic "other"
label), describe how the exporter/collector will emit metrics (only allowed
queue labels appear verbatim; disallowed names are hashed/bucketed or
aggregated), and include default config values and expected cardinality behavior
in the doc.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | PRD now documents `metrics.allowed_queues`, fallback hashing, and default cardinality controls. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> PRD now documents `metrics.allowed_queues`, fallback hashing, and default cardinality controls. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:682
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## `

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350574887

- [review_comment] 2025-09-16T03:18:41Z by coderabbitai[bot] (.claude/agents/feature-enhancer.md:12)

```text
`
In .claude/agents/feature-enhancer.md around lines 8 to 12 (and similar issues
at 20-36, 44-74, 80-126, 190-212, 252-269), the markdown has lint failures
(incorrect heading spacing, missing blank lines around fenced code blocks,
inconsistent fenced code ticks, trailing newline missing). Run make mdlint-fix
and commit the result; if fixing manually, ensure headings have a blank line
above, add a blank line before and after every
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Markdown headings now include required spacing and fenced blocks are separated properly. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Markdown headings now include required spacing and fenced blocks are separated properly. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:722
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In AGENTS.md around lines 183 to 193, replace the fake placeholder link

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350574894

- [review_comment] 2025-09-16T03:18:41Z by coderabbitai[bot] (AGENTS.md:193)

```text
In AGENTS.md around lines 183 to 193, replace the fake placeholder link
"https://fake.com" and any "(to PR)" text with the actual PR or issue URLs (or
remove them entirely if no real links exist), and update the nonstandard
admonition syntax "[!info]-" to the proper GitHub-flavored blockquote admonition
format such as "> [!NOTE]" or "> [!WARNING]" (without the trailing hyphen),
ensuring links use valid markdown link syntax and no placeholder text remains.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Replaced placeholder links and converted admonitions to GitHub-style syntax. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Replaced placeholder links and converted admonitions to GitHub-style syntax. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:759
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In cmd/tui/main.go around line 31, the FlagSet is created with flag.ExitOnError

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350574900

- [review_comment] 2025-09-16T03:18:42Z by coderabbitai[bot] (cmd/tui/main.go:31)

```text
In cmd/tui/main.go around line 31, the FlagSet is created with flag.ExitOnError
which calls os.Exit on parse errors and bypasses your function's error handling;
change the creation to use flag.ContinueOnError so Parse returns an error
instead, then call fs.Parse(os.Args[1:]) and handle/return that error
consistently with the rest of the function (optionally set fs.SetOutput to
discard or a logger if you want to suppress default printed usage).
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Flag parsing now uses `ContinueOnError` with explicit error handling. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Flag parsing now uses `ContinueOnError` with explicit error handling. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:795
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In cmd/tui/main.go around line 43, the error returned by fs.Parse(os.Args[1:])

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350574908

- [review_comment] 2025-09-16T03:18:42Z by coderabbitai[bot] (cmd/tui/main.go:43)

```text
In cmd/tui/main.go around line 43, the error returned by fs.Parse(os.Args[1:])
is being discarded; instead capture the error, check if it's non-nil, write the
error message to stderr (and/or call fs.Usage()), and exit with a non-zero
status (e.g., os.Exit(2>) or return the error) so parsing failures are properly
reported and the process terminates.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Parse errors now propagate and abort the program with a clear message. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Parse errors now propagate and abort the program with a clear message. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:831
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In cmd/tui/main.go around lines 64-66, the code pings Redis and merely logs on

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350574913

- [review_comment] 2025-09-16T03:18:42Z by coderabbitai[bot] (cmd/tui/main.go:66)

```text
In cmd/tui/main.go around lines 64-66, the code pings Redis and merely logs on
error but continues running with a broken connection; change this to fail fast:
on Ping error write the error to stderr (or process logger) and exit the process
with a non-zero status (or return the error from main) so the TUI does not
proceed with an invalid Redis client; alternatively implement a small
retry/backoff loop before exiting if transient errors are expected, but do not
allow execution to continue when Ping ultimately fails.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Redis ping now exits on failure and CLI flags map into config/TUI options. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Redis ping now exits on failure and CLI flags map into config/TUI options. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:866
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In cmd/tui/main.go around line 68, the TODO leaves many CLI flags un-wired

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350574918

- [review_comment] 2025-09-16T03:18:42Z by coderabbitai[bot] (cmd/tui/main.go:68)

```text
In cmd/tui/main.go around line 68, the TODO leaves many CLI flags un-wired
(redisURL, cluster, namespace, readOnly, metricsAddr, theme, fps) so the TUI
starts without required runtime options; fix by reading these flags from the
root/Cobra command or shared config struct and passing them into the TUI
initializer: add parameters on the TUI options/ctor for
redisURL/cluster/namespace (used to construct the redis/backend client),
readOnly (toggle input/editing and backend write operations), metricsAddr
(start/forward metrics collection if non-empty), and theme/fps (apply to
renderer/refresh loop); ensure the flags are defined and defaulted at command
setup and validate/convert types before constructing the TUI, then remove the
TODO.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | TUI receives structured options (cluster, namespace, read-only, theme, fps, metrics addr) instead of ignoring flags. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TUI receives structured options (cluster, namespace, read-only, theme, fps, metrics addr) instead of ignoring flags. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:903
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/admin-api/deploy.sh around lines 1 to 10, the script uses weak

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350574922

- [review_comment] 2025-09-16T03:18:42Z by coderabbitai[bot] (deployments/admin-api/deploy.sh:10)

```text
In deployments/admin-api/deploy.sh around lines 1 to 10, the script uses weak
bash settings and unquoted variable expansions; update the shebang area to
enable strict mode by replacing the current set -e with set -Eeuo pipefail, and
ensure all variable usages in the script are quoted (e.g., use "$NAMESPACE",
"$APP_NAME", "${ENVIRONMENT}", "${VERSION}" wherever referenced) to prevent
word-splitting and undefined-variable errors; also ensure positional defaults
remain as shown but are safely assigned (retain ENVIRONMENT="${1:-staging}" and
VERSION="${2:-latest}").
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Script now runs with `set -Eeuo pipefail`, uses a safe IFS, and quotes variables. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Script now runs with `set -Eeuo pipefail`, uses a safe IFS, and quotes variables. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:944
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/admin-api/deploy.sh around lines 31 to 43, the docker tag/push

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350574938

- [review_comment] 2025-09-16T03:18:42Z by coderabbitai[bot] (deployments/admin-api/deploy.sh:43)

```text
In deployments/admin-api/deploy.sh around lines 31 to 43, the docker tag/push
steps use unquoted variable expansions and assume a good registry value which
can break with spaces, empty vars, or missing namespace; update to quote all
variable expansions (e.g. "${APP_NAME}" "${VERSION}" "${REGISTRY_URL}" ),
validate or default DOCKER_REGISTRY and DOCKER_NAMESPACE explicitly before use
(fail fast if missing), construct a fully qualified repository string like
"${REGISTRY_URL%/}/${DOCKER_NAMESPACE:-your-namespace}/${APP_NAME}:${VERSION}"
to avoid double slashes or implicit Docker Hub quirks, and use that quoted
repository value for both docker tag and docker push; also consider requiring
docker login before push (or check LOGIN env) and exit with an error if push
fails.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Image tagging/pushing now validates registry inputs and uses a normalized repository string. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Image tagging/pushing now validates registry inputs and uses a normalized repository string. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:982
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/admin-api/deploy.sh around lines 49 to 61, the script creates the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350574945

- [review_comment] 2025-09-16T03:18:43Z by coderabbitai[bot] (deployments/admin-api/deploy.sh:61)

```text
In deployments/admin-api/deploy.sh around lines 49 to 61, the script creates the
${NAMESPACE} but then applies manifests without specifying that namespace;
update the kubectl apply invocations to target the created namespace by adding
-n ${NAMESPACE} (or --namespace=${NAMESPACE}) to the redis and admin-api apply
commands so both kubectl apply -f deployments/admin-api/k8s-redis.yaml and
kubectl apply -f deployments/admin-api/k8s-deployment.yaml run against the
intended namespace; keep the existing kubectl wait which already uses -n
${NAMESPACE}.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | `kubectl apply` now targets the intended namespace for Redis and Admin API manifests. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> `kubectl apply` now targets the intended namespace for Redis and Admin API manifests. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:1023
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/admin-api/deploy.sh around lines 73-85 (and also apply same

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350574957

- [review_comment] 2025-09-16T03:18:43Z by coderabbitai[bot] (deployments/admin-api/deploy.sh:85)

```text
In deployments/admin-api/deploy.sh around lines 73-85 (and also apply same
change to lines 113-116), the script currently starts a kubectl port-forward in
the "local" branch without ensuring it is cleaned up and also runs port-forward
in Docker "local" mode; update the script so it only performs kubectl
port-forward when running against a k8s cluster (not when ENVIRONMENT indicates
Docker/local-Docker), capture the background PID immediately, set a trap on EXIT
(and SIGINT/SIGTERM) to kill the port-forward if it was started, and ensure any
sleep/wait uses a small readiness check rather than a blind sleep; apply the
identical guard+trap pattern to the other location (lines 113-116).
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Port-forwarding now runs only during k8s smoke tests and is cleaned up via trap. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Port-forwarding now runs only during k8s smoke tests and is cleaned up via trap. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:1061
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/admin-api/deploy.sh around lines 170-177, replace the deprecated

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350574963

- [review_comment] 2025-09-16T03:18:43Z by coderabbitai[bot] (deployments/admin-api/deploy.sh:177)

```text
In deployments/admin-api/deploy.sh around lines 170-177, replace the deprecated
docker-compose usage with the modern `docker compose` subcommand and add a
pre-check that verifies `docker compose` is available: implement a small check
(e.g., run `docker compose version` or `command -v docker` + `docker compose
version` and test its exit code) before the local deploy branch and if the check
fails print a clear error and exit non-zero; then change the local case to call
`docker compose -f deployments/admin-api/docker-compose.yaml up -d` (or bail if
unavailable).
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Local deploy now requires `docker compose` and invokes the modern subcommand. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Local deploy now requires `docker compose` and invokes the modern subcommand. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:1100
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/admin-api/docker-compose.yaml around line 28, the JWT_SECRET is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350574973

- [review_comment] 2025-09-16T03:18:43Z by coderabbitai[bot] (deployments/admin-api/docker-compose.yaml:28)

```text
In deployments/admin-api/docker-compose.yaml around line 28, the JWT_SECRET is
set to a weak default which is insecure; remove the hardcoded fallback so the
environment variable must be provided (e.g. change JWT_SECRET=${JWT_SECRET} with
no default), add a tiny validation service or init container that checks
process.env.JWT_SECRET and exits non‑zero if missing, and update the admin-api
service to depend_on that validation service so the stack fails fast when
JWT_SECRET is not supplied; ensure local/dev docs or .env.example instructs how
to provide a strong secret.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Compose stack now enforces `JWT_SECRET` via an env-check container and no longer ships a weak default. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Compose stack now enforces `JWT_SECRET` via an env-check container and no longer ships a weak default. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_004.md:1138
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/admin-api/docker-compose.yaml around line 55 the file is missing

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350574988

- [review_comment] 2025-09-16T03:18:43Z by coderabbitai[bot] (deployments/admin-api/docker-compose.yaml:55)

```text
In deployments/admin-api/docker-compose.yaml around line 55 the file is missing
a trailing newline at EOF; open the file and add a single newline character
after the last line ("driver: bridge") so the file ends with a newline, save and
commit the change.
```

> [!INFO]- **Rejected**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | No | Yes | 7 | Confirmed the compose file already ends with a newline, so no edit was required. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Reviewer note deemed non-actionable. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:28
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/docker/docker-compose.yaml around lines 36-38 (and also update

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350574996

- [review_comment] 2025-09-16T03:18:43Z by coderabbitai[bot] (deployments/docker/docker-compose.yaml:38)

```text
In deployments/docker/docker-compose.yaml around lines 36-38 (and also update
similar occurrences at 45-49 and 74-77), the RBAC mount is ./rbac-configs ->
/app/rbac but the env vars point to /app/configs/*.yaml and API_TOKEN_* values
are inlined; change the RBAC env vars to point to the actual mounted path
(/app/rbac/roles.yaml and /app/rbac/resources.yaml) and remove any hard-coded
API_TOKEN_* from the compose file, instead reference an env_file (admin-api.env)
or Docker secrets; update service definitions to load admin-api.env (ensure that
file is .gitignored) or declare docker secrets and replace API_TOKEN_* entries
with secret references.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Verified the compose stack now reads RBAC configs from the mounted path and loads API tokens via env_file. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Verified the compose stack now reads RBAC configs from the mounted path and loads API tokens via env_file. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:66
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/docker/docker-compose.yaml around lines 121 to 123, the file is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575000

- [review_comment] 2025-09-16T03:18:44Z by coderabbitai[bot] (deployments/docker/docker-compose.yaml:123)

```text
In deployments/docker/docker-compose.yaml around lines 121 to 123, the file is
missing a trailing newline at EOF and the review flagged potential
incompatibility with depends_on health condition; add a single newline character
at the end of the file, and if you rely on depends_on: condition:
service_healthy ensure each dependent service defines a proper healthcheck and
verify behaviour against your Compose implementation (run docker compose
version) — if incompatibility exists, either remove the condition and gate
startup in the service entrypoint or migrate to a Compose version that supports
health conditions.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Compose v2.39.2 confirmed to honour service_healthy gating and the manifest now carries the required newline. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Compose v2.39.2 confirmed to honour service_healthy gating and the manifest now carries the required newline. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:105
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/docker/Dockerfile.admin-api around lines 52-53, the HEALTHCHECK

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575007

- [review_comment] 2025-09-16T03:18:44Z by coderabbitai[bot] (deployments/docker/Dockerfile.admin-api:53)

```text
In deployments/docker/Dockerfile.admin-api around lines 52-53, the HEALTHCHECK
uses wget which is inconsistent with other Dockerfiles; replace the wget
invocation with an equivalent curl-based healthcheck command (use curl options
to fail on non-2xx, be silent, set a timeout, and return non-zero on failure)
and ensure the Dockerfile installs curl in the runtime image (add curl to the
runtime package list or use a base image that includes curl) so the healthcheck
works and matches project convention.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Runtime layer now installs curl and the healthcheck aligns with other images. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Runtime layer now installs curl and the healthcheck aligns with other images. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:144
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/docker/Dockerfile.rbac-token-service around lines 39-40, the COPY

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575011

- [review_comment] 2025-09-16T03:18:44Z by coderabbitai[bot] (deployments/docker/Dockerfile.rbac-token-service:40)

```text
In deployments/docker/Dockerfile.rbac-token-service around lines 39-40, the COPY
commands currently suppress failures with "2>/dev/null || true", which hides
missing config files; decide whether these configs are required or optional and
remove the error-suppression accordingly: if required, delete the redirections
and "|| true" so the build fails loudly when files are missing and update build
artifacts to include the files; if optional, remove the suppressions and instead
implement an explicit presence check (or adjust the application to handle
missing configs) and use a conditional copy strategy in the Docker build
pipeline (e.g., stage the files only when present) so missing files are handled
intentionally rather than silently ignored.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Build now fails loudly if RBAC config bundle is missing. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Build now fails loudly if RBAC config bundle is missing. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:181
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/docker/rbac-configs/resources.yaml around line 231, the file is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575018

- [review_comment] 2025-09-16T03:18:44Z by coderabbitai[bot] (deployments/docker/rbac-configs/resources.yaml:231)

```text
In deployments/docker/rbac-configs/resources.yaml around line 231, the file is
missing a trailing newline; add a single newline character at the end of the
file so the last line ("GET /api/v1/admin/audit": ["admin:audit"]) is terminated
by a newline to satisfy linters and POSIX file conventions.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Endpoint map now mirrors the Admin API contract and includes metrics/health routes. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Endpoint map now mirrors the Admin API contract and includes metrics/health routes. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:142
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/docker/rbac-configs/roles.yaml around lines 95 to 102, add a

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575024

- [review_comment] 2025-09-16T03:18:44Z by coderabbitai[bot] (deployments/docker/rbac-configs/roles.yaml:102)

```text
In deployments/docker/rbac-configs/roles.yaml around lines 95 to 102, add a
terminating newline at EOF to satisfy YAML linters, and update the file to
explicitly document how role_assignment_rules are enforced and their precedence:
state which component (token service, API gateway, or other) applies these
rules, describe whether matching is based on token claims, user attributes, or
group memberships, and clarify precedence order between domain-based rules and
explicit user/group mappings (e.g., user/group overrides domain rules or vice
versa). Ensure the documentation snippet is concise, placed near the
role_assignment_rules section, and includes an example showing resolution order
so readers know which mapping wins.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Comments now explain how domain rules interact with explicit mappings and newline is restored. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Comments now explain how domain rules interact with explicit mappings and newline is restored. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:255
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/docker/rbac-configs/token-service.yaml around line 28, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575032

- [review_comment] 2025-09-16T03:18:44Z by coderabbitai[bot] (deployments/docker/rbac-configs/token-service.yaml:28)

```text
In deployments/docker/rbac-configs/token-service.yaml around line 28, the
encryption_key currently falls back to the insecure default
"default-dev-key-change-in-production"; remove that hardcoded default so the
value must come from RBAC_KEY_ENCRYPTION_KEY (e.g. use the env var without a
default), and add startup validation that requires RBAC_KEY_ENCRYPTION_KEY (and
REDIS_PASSWORD) and fails fast on missing vars; implement a small
startup_validation section that lists required_env_vars and fail_fast: true so
the service refuses to start when those secrets are not provided.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Config now insists on external secrets and bails if they are missing. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Config now insists on external secrets and bails if they are missing. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:295
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/docker/rbac-configs/token-service.yaml around lines 72-75 the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575036

- [review_comment] 2025-09-16T03:18:44Z by coderabbitai[bot] (deployments/docker/rbac-configs/token-service.yaml:75)

```text
In deployments/docker/rbac-configs/token-service.yaml around lines 72-75 the
CORS allowed_origins contains hardcoded staging.example.com and example.com
domains which are environment-specific; replace these hardcoded entries with a
generated list sourced from an environment variable (e.g., ALLOWED_ORIGINS) or a
secret/config map so the manifest is templated at deploy time; update the
manifest/template to read a comma-separated or YAML/JSON list from the env var
(or ConfigMap) and document the required env var format so different
environments can supply their own allowed origins.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | CORS origins now come from env vars, eliminating hardcoded staging/prod hosts. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> CORS origins now come from env vars, eliminating hardcoded staging/prod hosts. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:333
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/docker/rbac-configs/token-service.yaml around line 114, the file

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575042

- [review_comment] 2025-09-16T03:18:45Z by coderabbitai[bot] (deployments/docker/rbac-configs/token-service.yaml:114)

```text
In deployments/docker/rbac-configs/token-service.yaml around line 114, the file
is missing a trailing newline; add a single newline character at the end of the
file so the final line ("content_security_policy: \"default-src 'self'\"") is
terminated with a proper line ending and save the file to ensure POSIX-compliant
EOF newline is present.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 10 | File now ends with a proper newline as required by linters. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> File now ends with a proper newline as required by linters. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:371
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/README-RBAC-Deployment.md around lines 19 to 36, replace the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575049

- [review_comment] 2025-09-16T03:18:45Z by coderabbitai[bot] (deployments/README-RBAC-Deployment.md:36)

```text
In deployments/README-RBAC-Deployment.md around lines 19 to 36, replace the
ASCII art block with a precise architecture section: remove the diagram and
instead add clear, textual and/or tool-compatible diagram content that documents
network topology (node/pod/service relationships and ingress/egress paths),
security boundaries (namespaces, network policies, RBAC roles and token flows),
data flow diagrams (request/response paths between Admin API, RBAC Token
Service, Redis and monitoring stack), exact port configurations (service and
container ports, protocols), and explicit service dependencies and startup
order; if a visual diagram is required, include a PlantUML or Mermaid diagram
snippet and link to any external draw.io/diagrams file so the doc is
machine-readable and appropriate for production use.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | ASCII art replaced with a detailed, tool-friendly architecture walkthrough. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> ASCII art replaced with a detailed, tool-friendly architecture walkthrough. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:406
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/README-RBAC-Deployment.md around lines 138-146, the env var table

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575055

- [review_comment] 2025-09-16T03:18:45Z by coderabbitai[bot] (deployments/README-RBAC-Deployment.md:146)

```text
In deployments/README-RBAC-Deployment.md around lines 138-146, the env var table
exposes secret defaults (Redis password, RBAC_SIGNING_KEY, RBAC_ENCRYPTION_KEY);
remove any hint of actual values and replace the Default column for these
entries with a neutral placeholder such as "REDACTED — provide via secret" or
"must be set from a secret", and add a short note (one line) on how to reference
them (e.g., via Kubernetes Secret, Docker secret, or env var injection) without
printing values; do not include actual keys or passwords anywhere in the README.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Table now points deployers to secrets instead of leaking representative values. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Table now points deployers to secrets instead of leaking representative values. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:447
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/README-RBAC-Deployment.md around lines 257 to 266, the "STOP at

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575064

- [review_comment] 2025-09-16T03:18:45Z by coderabbitai[bot] (deployments/README-RBAC-Deployment.md:266)

```text
In deployments/README-RBAC-Deployment.md around lines 257 to 266, the "STOP at
staging" note leaves the production deployment section incomplete; replace it
with a full production deployment procedure (or remove the section if not
intended) by adding concrete, ordered steps covering: blue‑green (or canary)
deployment workflow and tooling, detailed rollback and emergency rollback
procedures, database migration plan with safe migration strategy and backups,
production-specific security hardening and approvals, load balancing and scaling
configuration, load and smoke testing acceptance criteria, monitoring/alerting
runbook and SLOs, disaster recovery and backup verification steps, and required
pre/post‑deployment checklist and signoff process so the file contains an
actionable, auditable production deployment guide.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Production section now contains a full, auditable runbook. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Production section now contains a full, auditable runbook. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:484
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/advanced-rate-limiting-api.md around lines 140–155, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575072

- [review_comment] 2025-09-16T03:18:45Z by coderabbitai[bot] (docs/api/advanced-rate-limiting-api.md:155)

```text
In docs/api/advanced-rate-limiting-api.md around lines 140–155, the
FairnessConfig docs lack a guardrail explaining that cumulative minimum shares
can exceed 100%; update the docs to state the constraint "Sum(Weights) *
MinGuaranteedShare must be < 1.0" and add a clear fallback behavior: when the
constraint is violated the implementation will renormalize and cap minimum
guarantees (proportionally scale down per-priority MinGuaranteedShare so the
total < 1.0 and ensure no negative/greater-than-1 values), and mention any
rounding or epsilon used; keep the wording concise and add an example or single
sentence showing renormalization occurs when exceeded.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Docs now spell out the MinGuaranteedShare guardrail and fallback behaviour. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Docs now spell out the MinGuaranteedShare guardrail and fallback behaviour. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:525
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/advanced-rate-limiting-api.md around lines 374 to 381, the note that

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575084

- [review_comment] 2025-09-16T03:18:46Z by coderabbitai[bot] (docs/api/advanced-rate-limiting-api.md:383)

```text
In docs/api/advanced-rate-limiting-api.md around lines 374 to 381, the note that
each consume is a "single Redis round‑trip via Lua" lacks durability and Redis
Cluster details; add a "Redis Details" subsection that (1) instructs preloading
the Lua script (SCRIPT LOAD) and using EVALSHA with a safe fallback to EVAL on
NOSCRIPT, (2) documents key slotting requirements for Redis Cluster and
recommends a hash‑tag naming convention (example pattern like
{rl}:{scope}:bucket) so all keys share the same slot, and (3) specifies
operational guidance for handling transient Redis errors: timeouts, exponential
backoff and retry on NOSCRIPT, handling READONLY errors during failover/replica
writes, and suggested retry limits and logging for observability.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/anomaly-radar-slo-budget.md around lines 36 to 38, the Go import

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575092

- [review_comment] 2025-09-16T03:18:46Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:38)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 36 to 38, the Go import
uses the path
"github.com/flyingrobots/go-redis-work-queue/internal/anomaly-radar-slo-budget"
but the code references the package as anomalyradarslobudget which will not
compile; update the import to use an explicit alias (e.g. anomalyradarslobudget
"github.com/flyingrobots/go-redis-work-queue/internal/anomaly-radar-slo-budget")
and ensure the target package file declares package anomalyradarslobudget
(rename the package identifier if it currently contains hyphens or a different
name).
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Added a public re-export at `pkg/anomaly-radar-slo-budget` and updated docs to reference the stable path. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added a public re-export at `pkg/anomaly-radar-slo-budget` and updated docs to reference the stable path. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:637
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/calendar-view.md around lines 39 to 45, the CalendarView struct uses

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575102

- [review_comment] 2025-09-16T03:18:46Z by coderabbitai[bot] (docs/api/calendar-view.md:45)

```text
In docs/api/calendar-view.md around lines 39 to 45, the CalendarView struct uses
*time.Location for the timezone which does not JSON-marshal as a simple string;
change the Timezone field to a string (e.g., Timezone string `json:"timezone"` )
and document/expect an IANA timezone name (or offset) in requests/responses,
then update the Go example below to use Timezone: "UTC" (or another string)
instead of Timezone: "UTC" as a *time.Location value so clients receive a plain
string.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Calendar view now expects string timezones, matching JSON clients. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Calendar view now expects string timezones, matching JSON clients. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:643
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/calendar-view.md around lines 193 to 241, the bulk reschedule

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575110

- [review_comment] 2025-09-16T03:18:46Z by coderabbitai[bot] (docs/api/calendar-view.md:241)

```text
In docs/api/calendar-view.md around lines 193 to 241, the bulk reschedule
endpoint lacks guardrails and transactional behavior details; update the docs to
state a configurable max items per bulk request (default 100), describe
server-side rate limit expectations, and require clients to paginate or batch
larger payloads; specify that the response returns an overall transaction id and
a per-item status array (success/failure with error reasons), and document how
partial failures are handled (best-effort rollback when feasible, otherwise
leave successful items committed and mark failed ones with retry guidance) so
clients can implement safe retry/backoff.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Bulk reschedule docs now define limits, transaction IDs, and partial failure semantics. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Bulk reschedule docs now define limits, transaction IDs, and partial failure semantics. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:680
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/calendar-view.md around lines 575 to 582, the table exposes numeric

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575116

- [review_comment] 2025-09-16T03:18:46Z by coderabbitai[bot] (docs/api/calendar-view.md:582)

```text
In docs/api/calendar-view.md around lines 575 to 582, the table exposes numeric
enum values in public JSON; change the public representation to use string
status values instead of numbers and document the allowed string set (scheduled,
running, completed, failed, canceled). Update the table and any JSON examples to
show "status":"scheduled"|"running"|"completed"|"failed"|"canceled" and add a
note that the server may map these strings to internal integer constants. Ensure
examples and descriptions consistently use the string names and remove raw
numeric examples.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Calendar view docs now use string statuses/view types to match JSON payloads. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Calendar view docs now use string statuses/view types to match JSON payloads. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:719
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/distributed-tracing-integration.md around lines 41 to 50, the sample

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575123

- [review_comment] 2025-09-16T03:18:46Z by coderabbitai[bot] (docs/api/distributed-tracing-integration.md:50)

```text
In docs/api/distributed-tracing-integration.md around lines 41 to 50, the sample
YAML sets insecure: true which encourages unsafe production usage; change the
example to show insecure: false by default and demonstrate TLS first (e.g., use
otlp+https or a TLS endpoint URL and set insecure: false), and update the text
to flip the recommendation order (TLS first, then opt-in insecure for local
dev). Additionally add a short note stating the system should fail fast when the
configured endpoint appears to require TLS (e.g., uses https or known TLS port)
but insecure: true is set, instructing implementers to validate endpoint
scheme/port and return a clear error if that mismatch is detected.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Default tracing config now prefers TLS and documents the safety check. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Default tracing config now prefers TLS and documents the safety check. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:757
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/distributed-tracing-integration.md around lines 348–359, the example

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575137

- [review_comment] 2025-09-16T03:18:47Z by coderabbitai[bot] (docs/api/distributed-tracing-integration.md:359)

```text
In docs/api/distributed-tracing-integration.md around lines 348–359, the example
shell commands use macOS-only tools (pbcopy, open); update the doc to either
detect and show cross-platform equivalents or explicitly flag the OS: for
copying, mention pbcopy (macOS), xclip/xsel (Linux) and clip or PowerShell
alternatives (Windows); for opening URLs, mention open (macOS), xdg-open (Linux)
and start or PowerShell Invoke-Item (Windows); add a short note explaining how
to detect OS in a script (or that users should pick the command for their OS) so
readers on Linux/Windows can follow the steps.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Added cross-platform clipboard/URL commands so instructions work beyond macOS. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added cross-platform clipboard/URL commands so instructions work beyond macOS. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:796
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/PRD.md around lines 130-131, the rate-limiter note is inaccurate:

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575159

- [review_comment] 2025-09-16T03:18:47Z by coderabbitai[bot] (docs/PRD.md:134)

```text
In docs/PRD.md around lines 130-131, the rate-limiter note is inaccurate:
calling INCR + EX=1s implements a fixed-window counter that allows bursts at
window boundaries; either amend the text to explicitly state this fixed-window
behavior and its bursty edge-case, or change the described implementation to a
Lua-based token-bucket (which is already referenced elsewhere in this PR) and
link to that section. Update the doc to clearly state which approach is used,
describe its observable behavior (bursting vs. smooth token refill), and, if
switching to the Lua token-bucket, point readers to the existing token-bucket
snippet/section in this PR for implementation details.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/PRD.md around lines 134-136, the current recommendation of using

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575173

- [review_comment] 2025-09-16T03:18:48Z by coderabbitai[bot] (docs/PRD.md:139)

```text
In docs/PRD.md around lines 134-136, the current recommendation of using
BRPOPLPUSH with a 1s per-queue timeout is latency-hostile for low-priority jobs;
update the doc to describe two configurable modes: (1) a low-latency mode that
reduces per-queue timeout to a much smaller value (e.g., 50-200ms) and explains
the increased CPU/redis load tradeoff, and (2) an atomic-priority mode that uses
a Lua script to probe priority queues and atomically RPOPLPUSH a job into
processing in one call (or a batched probe) and documents its complexity and
guarantees; add config knobs (mode name and timeout) to the spec and a short
paragraph comparing tradeoffs, recommended defaults, and when to choose each
mode.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/PRD.md around lines 169 to 174, the PRD currently omits a

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575188

- [review_comment] 2025-09-16T03:18:48Z by coderabbitai[bot] (docs/PRD.md:188)

```text
In docs/PRD.md around lines 169 to 174, the PRD currently omits a
Kubernetes-ready readiness endpoint; add a /readyz readiness probe now and
document the exact checks and expected responses. Implement a /readyz HTTP
endpoint in the service that performs: a Redis ping (fail if unreachable or auth
fails), verification that required worker goroutines have started and are
processing (e.g., heartbeat or running flag), and the circuit-breaker state
check (fail if open or tripped); return 200 with JSON {status:"ok",
checks:{...}} when all pass and 500 with details when any fail. Update the docs
to list each check, the exact probe path, expected JSON schema, and include
example k8s readinessProbe snippet (httpGet path:/readyz,
initialDelaySeconds/periodSeconds) so deployments can use it immediately.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## docs/SLAPS/worker-reflections/claude-001-reflection.md lines 39-43: the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575201

- [review_comment] 2025-09-16T03:18:49Z by coderabbitai[bot] (docs/SLAPS/worker-reflections/claude-001-reflection.md:43)

```text
docs/SLAPS/worker-reflections/claude-001-reflection.md lines 39-43: the
“Metrics-Disabled Testing” section is too vague — replace the hand‑waving with
an exact description of the toggle and patterns used to avoid Prometheus
duplicate-collector panics: state the config key or env var name (e.g.,
METRICS_ENABLED=false), describe the code guard used (e.g., wrap MustRegister
calls with a check or use a test-only conditional registry), and list the
concrete alternatives tried (per-test Registry via prometheus.NewRegistry(),
using promtest helpers, or using registerOnce/guarded register functions).
Include the exact snippet locations where tests should set the switch (e.g.,
test setup/teardown) and a short note on preferred approach for the repo.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Reflection now documents the exact metrics toggle, registry strategy, and alternatives tested. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Reflection now documents the exact metrics toggle, registry strategy, and alternatives tested. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:954
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In Makefile around lines 35 to 40, there is no clean target; add a PHONY clean

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350575225

- [review_comment] 2025-09-16T03:18:49Z by coderabbitai[bot] (Makefile:53)

```text
In Makefile around lines 35 to 40, there is no clean target; add a PHONY clean
target that removes common build artifacts and temporary files (e.g., build/,
dist/, *.o, *.pyc, .cache, node_modules/ or other project-specific outputs) and
update the .PHONY declaration to include clean so make clean always runs;
implement the clean rule to use safe rm -rf on those paths and keep it minimal
and project-appropriate.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/admin-api/deploy.sh around lines 123 to 137, the ServiceMonitor

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350577939

- [review_comment] 2025-09-16T03:19:56Z by coderabbitai[bot] (deployments/admin-api/deploy.sh:137)

```text
In deployments/admin-api/deploy.sh around lines 123 to 137, the ServiceMonitor
manifest uses the wrong apiVersion and assumes a named service port that may not
exist; change apiVersion from v1 to monitoring.coreos.com/v1 for the
ServiceMonitor CRD, and update your k8s Service definition (k8s-deployment.yaml)
to expose the port with name: http (e.g., name: http, port: <port-number>,
targetPort: <container-port>) so the ServiceMonitor endpoint port: http matches
a named port on the Service.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | ServiceMonitor now targets the correct CRD apiVersion. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> ServiceMonitor now targets the correct CRD apiVersion. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:1030
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/admin-api/k8s-deployment.yaml around lines 42-43 the jwt-secret

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350577957

- [review_comment] 2025-09-16T03:19:56Z by coderabbitai[bot] (deployments/admin-api/k8s-deployment.yaml:43)

```text
In deployments/admin-api/k8s-deployment.yaml around lines 42-43 the jwt-secret
is set to a weak hardcoded value ("change-me-in-production"); replace this by
generating a cryptographically secure secret (e.g., 32+ random bytes,
base64/hex-encoded) and remove the hardcoded value from the manifest, store the
secret in a Kubernetes Secret (or external secret manager like Vault/Secrets
Manager) and reference it via secretKeyRef/envFrom in the Deployment; ensure the
secret is not checked into source control, grant minimal RBAC to access it, and
document rotation/management procedures.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Secret manifest no longer ships a hard-coded JWT secret. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Secret manifest no longer ships a hard-coded JWT secret. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:1067
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/admin-api/k8s-deployment.yaml around line 65, the container image

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350577959

- [review_comment] 2025-09-16T03:19:56Z by coderabbitai[bot] (deployments/admin-api/k8s-deployment.yaml:65)

```text
In deployments/admin-api/k8s-deployment.yaml around line 65, the container image
is pinned to the non-deterministic tag "redis-work-queue/admin-api:latest";
replace this with a reproducible identifier (semantic version tag like
redis-work-queue/admin-api:vX.Y.Z or the image digest
redis-work-queue/admin-api@sha256:...) produced by your CI build, update the
manifest to use that tag/digest, and ensure your release pipeline publishes and
updates manifests automatically (or documents the manual step) so Kubernetes
deployments reference an immutable image for production.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Deployment manifest now references a tagged image instead of `latest`. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Deployment manifest now references a tagged image instead of `latest`. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:1105
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/admin-api/k8s-deployment.yaml around line 197, the file is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350577964

- [review_comment] 2025-09-16T03:19:56Z by coderabbitai[bot] (deployments/admin-api/k8s-deployment.yaml:197)

```text
In deployments/admin-api/k8s-deployment.yaml around line 197, the file is
missing a trailing newline at EOF; fix by adding a single newline character at
the end of the file so the file ends with a blank line, then commit;
additionally, add a CI linting rule (e.g., via eol-last in linters or a
YAML/file-format check) to enforce trailing newlines for all YAML files.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | YAML now ends with a newline and a lint script guards against regressions. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> YAML now ends with a newline and a lint script guards against regressions. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:1143
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/docker/docker-compose.yaml around lines 50 to 53, the healthcheck

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350577973

- [review_comment] 2025-09-16T03:19:57Z by coderabbitai[bot] (deployments/docker/docker-compose.yaml:53)

```text
In deployments/docker/docker-compose.yaml around lines 50 to 53, the healthcheck
uses wget which may not exist in minimal base images; replace it with a portable
curl invocation or run the check via CMD-SHELL so the shell builtin or installed
tools can be used. Update the test to use something like: use curl --fail
--silent --show-error http://localhost:8080/health (or wrap the existing wget in
a CMD-SHELL call) and keep interval/timeout/retries unchanged to ensure the
container healthcheck works across base images.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Healthcheck now uses curl with fail-fast flags, avoiding wget dependency. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Healthcheck now uses curl with fail-fast flags, avoiding wget dependency. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:28
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/docker/docker-compose.yaml around lines 86–106 (Prometheus) and

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350577980

- [review_comment] 2025-09-16T03:19:57Z by coderabbitai[bot] (deployments/docker/docker-compose.yaml:106)

```text
In deployments/docker/docker-compose.yaml around lines 86–106 (Prometheus) and
98–113 (Grafana), stop using :latest — pin Prometheus and Grafana images to
specific immutable versions (replace prom/prometheus:latest and
grafana/grafana:latest with chosen version tags), add a restart policy line
(restart: unless-stopped) to both service blocks, and resolve the missing
dashboards path referenced for Grafana by either adding the dashboards JSON
files under deployments/docker/grafana/dashboards or updating the Grafana
volumes entry to point to the correct existing dashboards directory.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Prometheus/Grafana now use pinned images, restart policies, and real provisioning files. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Prometheus/Grafana now use pinned images, restart policies, and real provisioning files. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:65
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/docker/rbac-configs/resources.yaml around lines 91 to 205 the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350577986

- [review_comment] 2025-09-16T03:19:57Z by coderabbitai[bot] (deployments/docker/rbac-configs/resources.yaml:205)

```text
In deployments/docker/rbac-configs/resources.yaml around lines 91 to 205 the
actions catalog is missing actions referenced by roles (metrics:read,
health:read) and contains an ambiguous queues:list entry; add explicit action
definitions for "metrics:read" and "health:read" mirroring the pattern
(description, risk_level, audit_required) used elsewhere, and resolve the
queues:list ambiguity by either removing "queues:list" from the catalog and all
roles or by explicitly mapping "queues:list" to the same internal capability as
"queues:read" (documenting it in the description) so semantics are not
duplicated or ambiguous.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Action catalog now covers metrics/health and documents the queues:list alias. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Action catalog now covers metrics/health and documents the queues:list alias. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:103
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## deployments/docker/rbac-configs/resources.yaml around lines 205–231: the DLQ

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350577993

- [review_comment] 2025-09-16T03:19:57Z by coderabbitai[bot] (deployments/docker/rbac-configs/resources.yaml:231)

```text
deployments/docker/rbac-configs/resources.yaml around lines 205–231: the DLQ
endpoints and verbs are mismatched with the Admin API; change the mapping key
"POST /api/v1/dlq/retry" to "POST /api/v1/dlq/requeue" and change the mapping
key "DELETE /api/v1/dlq" to "POST /api/v1/dlq/purge" while keeping "DELETE
/api/v1/queues/dlq": ["dlq:purge"] as-is; after making these edits, verify the
final endpoint→permission mappings against internal/admin-api/server.go (GET
/api/v1/dlq, POST /api/v1/dlq/requeue, POST /api/v1/dlq/purge, GET
/api/v1/workers) and internal/rbac-and-tokens/config.go, add monitoring
endpoints if missing (GET /metrics → metrics:read and GET /healthz →
health:read), and update related roles/tests/docs (test/e2e, test/integration,
docs/*, deployments/*) to reflect the canonical routes.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Endpoint map now mirrors the Admin API contract and includes metrics/health routes. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Endpoint map now mirrors the Admin API contract and includes metrics/health routes. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:142
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/docker/rbac-configs/roles.yaml around lines 19 to 23, the roles

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578007

- [review_comment] 2025-09-16T03:19:57Z by coderabbitai[bot] (deployments/docker/rbac-configs/roles.yaml:23)

```text
In deployments/docker/rbac-configs/roles.yaml around lines 19 to 23, the roles
file uses undefined actions and wildcards and uses ad-hoc condition flags that
don't match resources.yaml; update operator.permissions to replace "queues:list"
with "queues:read"; replace admin.permissions wildcard entries "admin:*" and
"system:*" with the explicit actions present in resources.yaml (e.g.,
"admin:users","admin:tokens","admin:audit","admin:system"); for
monitoring.permissions either add missing actions "metrics:read" and
"health:read" to deployments/docker/rbac-configs/resources.yaml or map them to
existing actions such as "stats:read"/"stats:export"; for emergency.permissions
replace "admin:all" and "emergency:*" with explicit actions (for example
admin:users, admin:tokens, admin:audit, admin:system, queues:delete, dlq:purge)
or implement true wildcard semantics in the enforcer (don’t leave magic); and
change emergency.resource_constraints.conditions to reference schedules defined
in resources.yaml (e.g., schedule: "emergency_only" or "after_hours") rather
than ad-hoc flags; finally run the RBAC validation to ensure every permission in
roles.yaml exists under actions in resources.yaml.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Roles now reference defined actions and list emergency powers explicitly. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Roles now reference defined actions and list emergency powers explicitly. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:183
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/kubernetes/admin-api-deployment.yaml around line 98, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578020

- [review_comment] 2025-09-16T03:19:57Z by coderabbitai[bot] (deployments/kubernetes/admin-api-deployment.yaml:98)

```text
In deployments/kubernetes/admin-api-deployment.yaml around line 98, the
container image is pinned to the non-deterministic tag
"work-queue/admin-api:latest"; replace it with a specific immutable tag
(semantic version like work-queue/admin-api:v1.2.3 or an image digest like
work-queue/admin-api@sha256:<digest>) so deployments are reproducible and not
affected by upstream image updates—update the manifest to point to the chosen
version/digest and ensure your CI/CD publishes and references that exact
tag/digest.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Deployment now references the tagged admin-api image. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Deployment now references the tagged admin-api image. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:229
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/kubernetes/admin-api-deployment.yaml around lines 190-193 the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578025

- [review_comment] 2025-09-16T03:19:58Z by coderabbitai[bot] (deployments/kubernetes/admin-api-deployment.yaml:193)

```text
In deployments/kubernetes/admin-api-deployment.yaml around lines 190-193 the
RoleRule grants get/list/watch on all configmaps and secrets which is too broad;
replace the blanket resource access with explicit resourceNames for each secret
and configmap the admin API actually needs (e.g. add resourceNames:
["<specific-secret-name>","<specific-configmap-name>"]) and remove wide-scoped
entries, or split into separate rules per resource type with only the minimal
verbs required; ensure the role is also scoped to the correct namespace and
update any references to match the new names.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | RBAC now grants get access only to the needed configmap/secret. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> RBAC now grants get access only to the needed configmap/secret. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:267
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/kubernetes/admin-api-deployment.yaml around line 271, the file is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578028

- [review_comment] 2025-09-16T03:19:58Z by coderabbitai[bot] (deployments/kubernetes/admin-api-deployment.yaml:271)

```text
In deployments/kubernetes/admin-api-deployment.yaml around line 271, the file is
missing a trailing newline; edit the file to add a single newline character at
the end (ensure the final line ends with a newline and save the file) so YAML
parsers and Git tools handle it correctly.
```

> [!INFO]- **Rejected**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | No | Yes | 8 | Verified the manifest already ends with a trailing newline. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Reviewer note deemed non-actionable. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:305
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/kubernetes/rbac-monitoring.yaml around lines 1 to 18, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578037

- [review_comment] 2025-09-16T03:19:58Z by coderabbitai[bot] (deployments/kubernetes/rbac-monitoring.yaml:18)

```text
In deployments/kubernetes/rbac-monitoring.yaml around lines 1 to 18, the
manifest uses a core API version for a ServiceMonitor (which is part of the
Prometheus Operator CRDs) and will fail; change the apiVersion to
monitoring.coreos.com/v1, keep kind: ServiceMonitor, ensure the ServiceMonitor
CRD is installed (Prometheus Operator) in the cluster and that the namespace
exists, and confirm the selector/labels match the target Service so Prometheus
can discover it.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | ServiceMonitor manifest now references monitoring.coreos.com/v1. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> ServiceMonitor manifest now references monitoring.coreos.com/v1. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:343
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/kubernetes/rbac-monitoring.yaml around line 43, the runbook_url

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578051

- [review_comment] 2025-09-16T03:19:58Z by coderabbitai[bot] (deployments/kubernetes/rbac-monitoring.yaml:43)

```text
In deployments/kubernetes/rbac-monitoring.yaml around line 43, the runbook_url
is pointing to a dummy/non-existent wiki; replace that value with the correct,
accessible runbook URL for the RBAC service (or remove the runbook_url field if
no runbook exists) so on-call engineers have a valid link. Ensure the new URL
points to the canonical incident runbook (or the team's runbook index) and
verify accessibility before committing.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Runbook link now points to the documented ops guide. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Runbook link now points to the documented ops guide. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:380
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/kubernetes/rbac-monitoring.yaml around lines 354 and 362, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578062

- [review_comment] 2025-09-16T03:19:58Z by coderabbitai[bot] (deployments/kubernetes/rbac-monitoring.yaml:354)

```text
In deployments/kubernetes/rbac-monitoring.yaml around lines 354 and 362, the
Slack webhook entries are set to the placeholder 'YOUR_SLACK_WEBHOOK_URL' which
will cause Alertmanager to fail or silently drop alerts; replace those
placeholders with the real Slack webhook URLs (or remove the Slack receiver
blocks entirely if you don't want Slack notifications), and instead of
hardcoding secrets inline ensure you reference the webhook via a Kubernetes
Secret (create a secret containing the webhook and use secretKeyRef or envFrom
in the Alertmanager config/Secret) so the actual webhook is not stored in plain
text in the manifest.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Slack receivers now read webhooks from mounted secrets instead of inline placeholders. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Slack receivers now read webhooks from mounted secrets instead of inline placeholders. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:416
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/kubernetes/rbac-monitoring.yaml around line 373, the file is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578065

- [review_comment] 2025-09-16T03:19:58Z by coderabbitai[bot] (deployments/kubernetes/rbac-monitoring.yaml:373)

```text
In deployments/kubernetes/rbac-monitoring.yaml around line 373, the file is
missing a trailing newline at EOF which can break YAML parsers; fix it by adding
a single newline character at the end of the file so the last line ("     
equal: [service, instance]") is terminated with a newline (ensure the file ends
with '\n').
```

> [!INFO]- **Rejected**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | No | Yes | 8 | Verified trailing newline already present after recent edits. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Reviewer note deemed non-actionable. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:455
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/kubernetes/rbac-token-service-deployment.yaml around lines

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578074

- [review_comment] 2025-09-16T03:19:59Z by coderabbitai[bot] (deployments/kubernetes/rbac-token-service-deployment.yaml:204)

```text
In deployments/kubernetes/rbac-token-service-deployment.yaml around lines
203-204, the container image is pinned to the "latest" tag with imagePullPolicy:
Always which causes non-deterministic deployments and rollback issues; replace
"work-queue/rbac-token-service:latest" with a semantic version tag (e.g. v1.2.3)
and set imagePullPolicy to IfNotPresent (or keep Always only if you must always
pull unique digests), update your release/CI pipeline to stamp the manifest with
the specific image tag (or image digest) during build/release so every
deployment references an immutable image and next time you need to roll back you
can reference the previous version.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Token service deployment now points to a tagged image and avoids Always pulls. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Token service deployment now points to a tagged image and avoids Always pulls. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:494
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/kubernetes/rbac-token-service-deployment.yaml around lines 301 to

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578076

- [review_comment] 2025-09-16T03:19:59Z by coderabbitai[bot] (deployments/kubernetes/rbac-token-service-deployment.yaml:307)

```text
In deployments/kubernetes/rbac-token-service-deployment.yaml around lines 301 to
307, the Role grants "watch" on configmaps and secrets which is excessive for a
token service; remove "watch" from the verbs for those resources (leaving "get"
and "list") unless you actually implement runtime reload behavior that relies on
watches — if you do need watches, document the exact code path and limit the
Role to the smallest namespace/specific resource names possible and add a
comment explaining why "watch" is required.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Removed unnecessary watch verb from configmap/secret access. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Removed unnecessary watch verb from configmap/secret access. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:533
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/kubernetes/rbac-token-service-deployment.yaml around lines 403 to

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578080

- [review_comment] 2025-09-16T03:19:59Z by coderabbitai[bot] (deployments/kubernetes/rbac-token-service-deployment.yaml:412)

```text
In deployments/kubernetes/rbac-token-service-deployment.yaml around lines 403 to
412, the file is missing a trailing newline at EOF; open the file and ensure
there is a single newline character at the end of the file (save with your
editor or run a formatter) so the file ends with a newline.
```

> [!INFO]- **Rejected**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | No | Yes | 8 | File already ends with a newline; no edit required. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Reviewer note deemed non-actionable. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:570
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/scripts/deploy-rbac-staging.sh around lines 17-19 (and similarly

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578092

- [review_comment] 2025-09-16T03:19:59Z by coderabbitai[bot] (deployments/scripts/deploy-rbac-staging.sh:19)

```text
In deployments/scripts/deploy-rbac-staging.sh around lines 17-19 (and similarly
21-23, 25-27) the log functions are duplicated; extract them to a single shared
shell utility (e.g., deployments/scripts/utils/logging.sh), put color constants
and log(), info(), warn(), error() helpers there, make the file idempotent if
sourced, and ensure it sets no side-effects; then replace the copied function
blocks in each deployment script with a single source ../utils/logging.sh (use
the correct relative path), verify scripts still execute (make the utility
readable/executable as needed) and run CI/lint to confirm no path or naming
regressions.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Logging helpers now live in `lib/logging.sh` and the script sources them. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Logging helpers now live in `lib/logging.sh` and the script sources them. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:608
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/scripts/deploy-rbac-staging.sh around line 155, the SERVICE_IP

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578097

- [review_comment] 2025-09-16T03:19:59Z by coderabbitai[bot] (deployments/scripts/deploy-rbac-staging.sh:155)

```text
In deployments/scripts/deploy-rbac-staging.sh around line 155, the SERVICE_IP
variable is assigned but never used; remove the dead assignment line entirely
(or if intended to be used, reference SERVICE_IP where needed) so there is no
unused variable left in the script and re-run ShellCheck to ensure no other
issues.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Removed unused SERVICE_IP assignment before shellcheck flagged it. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Removed unused SERVICE_IP assignment before shellcheck flagged it. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:647
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/scripts/deploy-rbac-staging.sh around line 211, the unquoted

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578102

- [review_comment] 2025-09-16T03:19:59Z by coderabbitai[bot] (deployments/scripts/deploy-rbac-staging.sh:196)

```text
In deployments/scripts/deploy-rbac-staging.sh around line 211, the unquoted
variable in the kill invocation can cause word-splitting or glob expansion
issues; update the command to quote the variable (e.g., kill
"$PORT_FORWARD_PID") and optionally guard against empty values (e.g., test for
non-empty before kill) to satisfy ShellCheck and avoid runtime failures.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/scripts/deploy-staging.sh around line 71, the docker build

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578104

- [review_comment] 2025-09-16T03:19:59Z by coderabbitai[bot] (deployments/scripts/deploy-staging.sh:71)

```text
In deployments/scripts/deploy-staging.sh around line 71, the docker build
invocation uses an unquoted variable ($IMAGE_NAME) which will break if it
contains spaces or special characters; update the command to quote the variable
(e.g. -t "$IMAGE_NAME") and review other shell variables in the script to ensure
all are quoted similarly to avoid word-splitting and globbing issues.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Docker build now quotes the image tag, preventing shell word-splitting. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Docker build now quotes the image tag, preventing shell word-splitting. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:717
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/scripts/deploy-staging.sh around lines 73, 85, 99, and 122, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578110

- [review_comment] 2025-09-16T03:20:00Z by coderabbitai[bot] (deployments/scripts/deploy-staging.sh:73)

```text
In deployments/scripts/deploy-staging.sh around lines 73, 85, 99, and 122, the
script uses archaic checks like "if [ $? -ne 0 ]; then" after commands; replace
each pattern by testing the command directly (e.g., change "some_command; if [
$? -ne 0 ]; then" to "if ! some_command; then" or invert as appropriate for
success checks), making sure any command substitutions, pipes, or redirects are
preserved and that the conditional logic (success vs failure) remains the same
for each occurrence.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Replaced `$?` checks with direct command conditions for build/push. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Replaced `$?` checks with direct command conditions for build/push. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:752
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/scripts/deploy-staging.sh around line 83 (and also at lines 112

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578112

- [review_comment] 2025-09-16T03:20:00Z by coderabbitai[bot] (deployments/scripts/deploy-staging.sh:83)

```text
In deployments/scripts/deploy-staging.sh around line 83 (and also at lines 112
and 192), several unquoted shell variables like $IMAGE_NAME are being used;
update each occurrence to use quoted parameter expansion (e.g., "$IMAGE_NAME")
so values with spaces or special characters are handled correctly, and review
surrounding commands to ensure quotes are applied consistently for all variable
expansions and any necessary escaping is preserved.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Quoted kubectl arguments and hardened port-forward cleanup throughout the script. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Quoted kubectl arguments and hardened port-forward cleanup throughout the script. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:789
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/scripts/deploy-staging.sh around line 225, the conditional uses

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578116

- [review_comment] 2025-09-16T03:20:00Z by coderabbitai[bot] (deployments/scripts/deploy-staging.sh:225)

```text
In deployments/scripts/deploy-staging.sh around line 225, the conditional uses
the deprecated negated -z test ("if [ ! -z \"$INGRESS_HOST\" ];"); replace it
with the canonical non-empty test using -n (i.e., "if [ -n \"$INGRESS_HOST\"
];") to match the project's style and avoid deprecated negation; ensure the
variable remains quoted to preserve whitespace and exit status semantics.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Condition now uses `[[ -n "$INGRESS_HOST" ]]`, matching suggested style. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Condition now uses `[[ -n "$INGRESS_HOST" ]]`, matching suggested style. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:825
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/scripts/health-check-rbac.sh around line 238, the current command

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578120

- [review_comment] 2025-09-16T03:20:00Z by coderabbitai[bot] (deployments/scripts/health-check-rbac.sh:238)

```text
In deployments/scripts/health-check-rbac.sh around line 238, the current command
uses grep | wc -l to count "error" occurrences; replace that pipeline with grep
-i -c to count matches directly and keep the existing stderr suppression and
options, i.e., invoke kubectl logs ... 2>/dev/null | grep -i -c error and assign
the output to error_count to follow ShellCheck best practices and be more
efficient.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Log error counting now uses `grep -i -c`, eliminating the extra pipeline. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Log error counting now uses `grep -i -c`, eliminating the extra pipeline. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:860
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/scripts/setup-monitoring.sh around lines 17 to 31 the logging

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578130

- [review_comment] 2025-09-16T03:20:00Z by coderabbitai[bot] (deployments/scripts/setup-monitoring.sh:31)

```text
In deployments/scripts/setup-monitoring.sh around lines 17 to 31 the logging
functions (log, info, warn, error) are duplicated; extract these functions into
a new deployments/scripts/lib/logging.sh and replace the duplicate block here
with a single source statement (e.g. source "$(dirname "$0")/lib/logging.sh");
do the same in the other scripts that currently copy these functions, remove the
duplicates, and ensure lib/logging.sh is idempotent (guard against
double-sourcing or redefinition by checking a sentinel variable) and uses the
same color variables so all scripts keep identical behavior.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Monitoring script sources the shared logging helpers instead of duplicating them. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Monitoring script sources the shared logging helpers instead of duplicating them. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:896
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/scripts/setup-monitoring.sh around lines 132 to 180, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578133

- [review_comment] 2025-09-16T03:20:00Z by coderabbitai[bot] (deployments/scripts/setup-monitoring.sh:180)

```text
In deployments/scripts/setup-monitoring.sh around lines 132 to 180, the
Alertmanager YAML is using hardcoded placeholder values (e.g., localhost:587 and
admin@company.com) which will not work in production; update the script to
source SMTP host/port, from address, and alert recipient addresses from
environment variables or Kubernetes Secrets (fallback to no-default or clearly
fail), replace the hardcoded values in the generated alertmanager.yml with those
variables, and add a brief comment or README note next to the creation step
documenting that real values or secrets must be provided before deployment.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Alertmanager config now sources SMTP + recipients from env vars. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Alertmanager config now sources SMTP + recipients from env vars. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:934
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/scripts/test-staging-deployment.sh around lines 63-68 (and

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578141

- [review_comment] 2025-09-16T03:20:00Z by coderabbitai[bot] (deployments/scripts/test-staging-deployment.sh:68)

```text
In deployments/scripts/test-staging-deployment.sh around lines 63-68 (and
similarly at 71-76 and 79-84), add a prerequisite check for the jq binary using
the same pattern as the kubectl check: use command -v jq &> /dev/null to detect
presence, call add_test_result "jq-available" "PASS" "jq is available" when
found, otherwise call add_test_result "jq-available" "FAIL" "jq not found" and
set all_passed=false; keep messaging consistent with the existing kubectl check
and ensure the new check appears before any code that uses jq.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Prerequisite checks now verify jq before the script proceeds. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Prerequisite checks now verify jq before the script proceeds. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:972
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/scripts/test-staging-deployment.sh around lines 331-333, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578148

- [review_comment] 2025-09-16T03:20:01Z by coderabbitai[bot] (deployments/scripts/test-staging-deployment.sh:333)

```text
In deployments/scripts/test-staging-deployment.sh around lines 331-333, the
base64 decode pipeline can fail silently and return an empty/invalid
bootstrap_token; update the retrieval to enable strict pipe error detection
(e.g., set -o pipefail for the block or run commands sequentially), check the
exit status of kubectl and base64, validate the token is non-empty and valid,
and on failure emit a clear error to stderr and exit with a non-zero status so
the script does not proceed with an invalid token.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Bootstrap token retrieval now handles pipeline failures safely. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Bootstrap token retrieval now handles pipeline failures safely. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:1009
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/scripts/test-staging-deployment.sh around lines 451-453 (and

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578153

- [review_comment] 2025-09-16T03:20:01Z by coderabbitai[bot] (deployments/scripts/test-staging-deployment.sh:453)

```text
In deployments/scripts/test-staging-deployment.sh around lines 451-453 (and
similarly lines 461-463), the script currently compares volume mount paths
against hardcoded values like "/app/audit" and "/app/keys"; change this to
reference configurable variables or values derived from the deployment manifest:
introduce variables at the top of the script (e.g., AUDIT_MOUNT_PATH and
KEYS_MOUNT_PATH) or query the deployment spec (kubectl get deployment -o
jsonpath=...) to obtain the expected mount paths, then replace the hardcoded
string literals in the comparisons with those variables so the checks follow the
manifest/config rather than fixed paths.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Volume mount checks now read expected paths from the deployment spec. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Volume mount checks now read expected paths from the deployment spec. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:1046
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/FEATURE_ENHANCEMENT_AGENT_PROMPT.md around lines 32 to 35, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578161

- [review_comment] 2025-09-16T03:20:01Z by coderabbitai[bot] (docs/FEATURE_ENHANCEMENT_AGENT_PROMPT.md:35)

```text
In docs/FEATURE_ENHANCEMENT_AGENT_PROMPT.md around lines 32 to 35, the
admonition uses a non-standard collapsible syntax (`> [!note]- **🗣️ CLAUDE'S
THOUGHTS 💭**`) which may not render across Markdown processors; replace it with
standard Markdown (e.g., a heading or blockquote) or explicitly document the
required Markdown extension, and update the template to use a portable pattern
such as a level-3 heading "### 🗣️ CLAUDE'S THOUGHTS" or a standard blockquote
starting with "> **🗣️ CLAUDE'S THOUGHTS:**" so rendering is consistent.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Template now uses a standard heading + blockquote for Claude's thoughts. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Template now uses a standard heading + blockquote for Claude's thoughts. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:1085
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/FEATURE_ENHANCEMENT_AGENT_PROMPT.md around lines 156–180 the color

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578168

- [review_comment] 2025-09-16T03:20:01Z by coderabbitai[bot] (docs/FEATURE_ENHANCEMENT_AGENT_PROMPT.md:180)

```text
In docs/FEATURE_ENHANCEMENT_AGENT_PROMPT.md around lines 156–180 the color
palette is embedded as an XML-style comment block which is not machine-friendly
or easily reusable; extract the palette into a proper config (e.g.,
docs/colors.yml or docs/colors.json) listing named roles and hex values, replace
the XML comment in the markdown with either a compact markdown table pointing to
the new config file or a brief reference link, and ensure each entry includes a
clear role key (background/panel/border/text/button) and hex value so downstream
code/docs can import or reference it.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Palette now lives in `docs/colors.yml` with a summary table in the template. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Palette now lives in `docs/colors.yml` with a summary table in the template. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:1122
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/TUI/README.md around lines 15 to 21, the README shows verified images

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578180

- [review_comment] 2025-09-16T03:20:01Z by coderabbitai[bot] (docs/TUI/README.md:31)

```text
In docs/TUI/README.md around lines 15 to 21, the README shows verified images
but is missing a short "How to build/run TUI" snippet; add a brief section
titled "How to build/run TUI" that states the command to run the TUI (make
run-tui) and a one-line note that referenced SVGs are in docs/TUI/images and the
Makefile target is defined (Makefile:19–20), so users can build and run the
interface easily.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In Makefile around line 13, remove the unnecessary GO111MODULE=on prefix from

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578185

- [review_comment] 2025-09-16T03:20:01Z by coderabbitai[bot] (Makefile:13)

```text
In Makefile around line 13, remove the unnecessary GO111MODULE=on prefix from
the go build command and ensure the linker variable used in LDFLAGS exists:
either add a package-level variable in package main (e.g. in cmd/$(APP)/main.go
declare var version string) so -X main.version=$(VERSION) will succeed, or
change the -X flag to point to the actual package and symbol that holds the
version; then run the build to confirm the "Missing main.version symbol for -X"
error is resolved.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Dropped the unused GO111MODULE env and confirmed the linker flag resolves against the existing `main.version` symbol. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Dropped the unused GO111MODULE env and confirmed the linker flag resolves against the existing `main.version` symbol. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:64
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In Makefile around lines 16 to 18, the TUI build target uses a different go

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350578194

- [review_comment] 2025-09-16T03:20:01Z by coderabbitai[bot] (Makefile:18)

```text
In Makefile around lines 16 to 18, the TUI build target uses a different go
build invocation than the other build targets; make it consistent by using the
same variables and flags as the other builds: ensure the target creates the bin
directory if needed and invoke go build with the shared GOFLAGS and LDFLAGS
(e.g. use GO111MODULE=on go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o bin/tui
./cmd/tui) so the TUI is built with the same environment and linker flags as the
rest of the project.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Unified the TUI build recipe with the primary binary and added the missing version symbol so LDFLAGS apply cleanly. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Unified the TUI build recipe with the primary binary and added the missing version symbol so LDFLAGS apply cleanly. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:101
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In .githooks/pre-commit around lines 7 to 15, the hook assumes python3 exists

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350580859

- [review_comment] 2025-09-16T03:21:08Z by coderabbitai[bot] (.githooks/pre-commit:15)

```text
In .githooks/pre-commit around lines 7 to 15, the hook assumes python3 exists
and will fail if it's absent; update it to detect an available interpreter at
runtime by checking for "python3" first, then "python" (using quoted paths), and
if neither exists either exit with a clear error message or skip the progress
update gracefully; use the detected interpreter variable when invoking
scripts/update_progress.py and ensure git add only runs when the update
succeeded.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Hook now discovers python3/python dynamically and exits early if neither exists. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Hook now discovers python3/python dynamically and exits early if neither exists. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:138
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In .github/workflows/update-progress.yml lines 1-6, YAMLLint flagged the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350580871

- [review_comment] 2025-09-16T03:21:08Z by coderabbitai[bot] (.github/workflows/update-progress.yml:6)

```text
In .github/workflows/update-progress.yml lines 1-6, YAMLLint flagged the
reserved key and bracket spacing: wrap the reserved key "on" in quotes (e.g.
"on":) and tighten the branch list brackets by removing inner spaces and quoting
the value (e.g. branches: ['main']) so the YAML parses cleanly.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Quoted the reserved `on` key and normalized the branches list to satisfy YAML linters. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Quoted the reserved `on` key and normalized the branches list to satisfy YAML linters. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:175
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In .github/workflows/update-progress.yml around lines 8 to 10, the workflow

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350580880

- [review_comment] 2025-09-16T03:21:08Z by coderabbitai[bot] (.github/workflows/update-progress.yml:10)

```text
In .github/workflows/update-progress.yml around lines 8 to 10, the workflow
lacks concurrency control which allows overlapping runs on rapid pushes; add a
top-level concurrency block (e.g., concurrency: { group: 'update-progress-${{
github.ref }}', cancel-in-progress: true }) to serialize runs per branch/ref and
cancel any in-progress run when a new one starts; place this block at the same
indentation level as permissions and jobs.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Added a concurrency block so only one update-progress run executes per ref. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added a concurrency block so only one update-progress run executes per ref. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:209
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## .github/workflows/update-progress.yml lines 31-41: the workflow uses unguarded

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350580886

- [review_comment] 2025-09-16T03:21:08Z by coderabbitai[bot] (.github/workflows/update-progress.yml:41)

```text
.github/workflows/update-progress.yml lines 31-41: the workflow uses unguarded
git add and unquoted $GITHUB_OUTPUT redirections; update the script to first
check for the presence of the files (or use a safe add that won’t fail) before
running git add, and ensure all instances of >> $GITHUB_OUTPUT are changed to >>
"$GITHUB_OUTPUT" (quote the variable) so the redirection target is not
word-split; keep behavior the same otherwise (commit only when changes exist).
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Hardened the commit step to quote `$GITHUB_OUTPUT` and only add files that exist. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Hardened the commit step to quote `$GITHUB_OUTPUT` and only add files that exist. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:245
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## .github/workflows/update-progress.yml around lines 43-47 contains an extra

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350580891

- [review_comment] 2025-09-16T03:21:08Z by coderabbitai[bot] (.github/workflows/update-progress.yml:47)

```text
.github/workflows/update-progress.yml around lines 43-47 contains an extra
trailing blank line after the git push step; remove the empty line after the
final "git push" run block so the file ends immediately after the command (no
extra newline line), keeping YAML lint happy.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Simplified the push step to a single-line run so no stray blank line remains. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Simplified the push step to a single-line run so no stray blank line remains. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:281
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In cmd/job-queue-system/main.go around lines 85 to 92, the code always starts

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350580900

- [review_comment] 2025-09-16T03:21:09Z by coderabbitai[bot] (cmd/job-queue-system/main.go:92)

```text
In cmd/job-queue-system/main.go around lines 85 to 92, the code always starts
the observability HTTP server (metrics/healthz/readyz) even for role=admin; this
can cause unnecessary port conflicts for one-shot admin commands. Modify the
logic to only call obs.StartHTTPServer and defer its Shutdown when the
configured role is not "admin" (e.g., if cfg.Role != "admin" or a bool IsAdmin
check is false); otherwise skip starting the HTTP server entirely. Ensure
readyCheck remains defined for non-admin paths and that no server is started or
shutdown deferred when running as admin.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Admin executions now skip starting the HTTP metrics server to avoid port conflicts. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Admin executions now skip starting the HTTP metrics server to avoid port conflicts. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:315
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In cmd/job-queue-system/main.go around lines 112–114, the background metrics

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350580914

- [review_comment] 2025-09-16T03:21:09Z by coderabbitai[bot] (cmd/job-queue-system/main.go:114)

```text
In cmd/job-queue-system/main.go around lines 112–114, the background metrics
updater is started unconditionally for CLI/admin paths; guard the
obs.StartQueueLengthUpdater call so it is skipped when running the admin/CLI
mode. Change the code to check an existing config flag (e.g. cfg.Admin or
cfg.IsAdmin) or add a new flag like cfg.DisableBackgroundUpdaters/DisableMetrics
and only call obs.StartQueueLengthUpdater when that flag is false; ensure the
updater is not launched for admin CLI invocations.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Admin runs no longer spawn the background queue length updater. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Admin runs no longer spawn the background queue length updater. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:353
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In cmd/job-queue-system/main.go around lines 142 to 149, the admin handling is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350580918

- [review_comment] 2025-09-16T03:21:09Z by coderabbitai[bot] (cmd/job-queue-system/main.go:149)

```text
In cmd/job-queue-system/main.go around lines 142 to 149, the admin handling is
nested inside the switch's default branch which obscures flow and risks
accidental execution when new roles are added; refactor by branching early or
adding an explicit case for "admin" instead of placing it in default: move the
admin check before the switch or add a dedicated case "admin" that calls
runAdmin(...) and returns, and keep the default branch solely to log/exit on
unknown roles so control flow is clear and future role additions won't
accidentally run admin logic.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Refactored the role switch to give admin its own case and leave default for true unknowns. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Refactored the role switch to give admin its own case and leave default for true unknowns. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:390
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In cmd/job-queue-system/main.go around lines 187-188, the purge-all branch

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350580928

- [review_comment] 2025-09-16T03:21:09Z by coderabbitai[bot] (cmd/job-queue-system/main.go:188)

```text
In cmd/job-queue-system/main.go around lines 187-188, the purge-all branch
prints a human-readable string with fmt.Printf which is inconsistent with other
admin commands that emit machine-readable JSON; replace the fmt.Printf call with
code that outputs a JSON object (e.g. {"purged": n}) to stdout using the
encoding/json package (or fmt.Fprintf with a properly escaped JSON string) and
ensure it ends with a newline and returns the same exit path as other admin
outputs.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Purge-all now emits JSON matching other admin outputs. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Purge-all now emits JSON matching other admin outputs. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:428
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/kubernetes/rbac-monitoring.yaml around line 35 (and also review

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350580943

- [review_comment] 2025-09-16T03:21:09Z by coderabbitai[bot] (deployments/kubernetes/rbac-monitoring.yaml:35)

```text
In deployments/kubernetes/rbac-monitoring.yaml around line 35 (and also review
lines 78, 88, 99, 109, 120, 130, 140, 151, 161, 172, 182), the Prometheus alert
expressions reference non-existent rbac_* metrics and job="rbac-token-service";
either register the missing rbac_* metrics in the RBAC token service (create and
MustRegister prometheus metrics, instrument code, and expose them on /metrics)
or update the YAML alerts to reference actual metric names we already export
(e.g., jobs_*, backpressure_*, exactly_once_*) and/or ensure Prometheus scrape
configs include a target with job="rbac-token-service" so
up{job="rbac-token-service"} is valid; choose one approach and make consistent
changes across all listed lines.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Updated the RBAC monitoring stack to use existing http/app metrics and dropped unreachable token counters. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Updated the RBAC monitoring stack to use existing http/app metrics and dropped unreachable token counters. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:465
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In Dockerfile around line 3, the build stage uses golang:1.23 while CI/docs

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350580952

- [review_comment] 2025-09-16T03:21:09Z by coderabbitai[bot] (Dockerfile:3)

```text
In Dockerfile around line 3, the build stage uses golang:1.23 while CI/docs
require Go 1.25; update the builder image tag to golang:1.25 to align
toolchains. Edit the FROM line to use golang:1.25 (or the exact patch variant
your CI/doc specify) so the compile environment matches CI/docs, then rebuild
and run CI to verify no compatibility issues.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 10 | Bumped the builder image to Go 1.25 to match CI/tooling requirements. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Bumped the builder image to Go 1.25 to match CI/tooling requirements. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:505
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/15_promotion_checklists.md around lines 21–33, the promotion checklist

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350580966

- [review_comment] 2025-09-16T03:21:09Z by coderabbitai[bot] (docs/15_promotion_checklists.md:33)

```text
In docs/15_promotion_checklists.md around lines 21–33, the promotion checklist
table is missing OWNER and EXIT CRITERIA columns which prevents enforceable
gates; add two new columns "OWNER" and "EXIT CRITERIA" to the table header and
for each row populate a responsible owner (GitHub handle or team) and a clear,
testable exit criterion (e.g., "E2E pass 5× in CI", ">=80% coverage on admin &
obs", "alert rules merged + helm docs") and update any status cells accordingly
so CI can validate these criteria; also add a short note under the table
describing how CI will read these columns (OWNER format and exact phrasing
required for automated checks).
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Added Owner/Exit Criteria columns across promotion tables and documented CI expectations. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added Owner/Exit Criteria columns across promotion tables and documented CI expectations. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:540
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/admin-api.md around lines 41 to 44, the configuration shows a single

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350580975

- [review_comment] 2025-09-16T03:21:10Z by coderabbitai[bot] (docs/api/admin-api.md:44)

```text
In docs/api/admin-api.md around lines 41 to 44, the configuration shows a single
confirmation_phrase while the endpoints require different phrases (e.g.,
CONFIRM_DELETE_ALL) causing inconsistency; choose one approach and make docs and
code consistent: either document separate keys (e.g., dlq_confirmation_phrase
and purge_all_confirmation_phrase) and update the sample config and README to
list both keys, or change the endpoints to validate against the single
configured confirmation_phrase and update any endpoint docs/samples to reference
that single key; apply the chosen change across the docs and codebase so the
config keys and endpoint expectations match.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Split destructive confirmations into DLQ/purge-all phrases and documented the fallback. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Split destructive confirmations into DLQ/purge-all phrases and documented the fallback. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:579
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/admin-api.md around lines 106 to 132, the queue parameter

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350581001

- [review_comment] 2025-09-16T03:21:10Z by coderabbitai[bot] (docs/api/admin-api.md:132)

```text
In docs/api/admin-api.md around lines 106 to 132, the queue parameter
description is ambiguous about alias-to-Redis-key mappings; update the docs to
explicitly list accepted aliases (high, low, completed, dead_letter) and show
the exact Redis key each alias resolves to (or state that a full Redis key may
be provided), and reference the configuration fields that control those mappings
by name (worker.queues.* for priority queues, completed_list for completed jobs,
dead_letter_list for dead-letter queue). Mention accepted value formats (alias
or full Redis key) and provide a short example mapping table or inline examples
referencing the config keys so readers know where to change the mappings.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Clarified queue alias mapping with a table tied to the worker config fields. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Clarified queue alias mapping with a table tied to the worker config fields. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:618
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/admin-api.md around lines 260 to 268, the CORS guidance currently

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350581008

- [review_comment] 2025-09-16T03:21:10Z by coderabbitai[bot] (docs/api/admin-api.md:268)

```text
In docs/api/admin-api.md around lines 260 to 268, the CORS guidance currently
implies using cors_allow_origins: ["*"]; update the text to recommend an empty
list as the safe default and explicitly warn that using "*" is dangerous when
require_auth: true. Replace the current bullet with instructions to set
cors_allow_origins to an explicit, environment-specific list of allowed origins
(or leave empty to block cross-origin requests), add a short note discouraging
"*" for authenticated endpoints, and include a brief recommendation to use
specific subdomains or environment variables for allowed origins and to test
CORS in staging before production.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Updated CORS guidance to promote explicit allow-lists and warn against `"*"` with auth. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Updated CORS guidance to promote explicit allow-lists and warn against `"*"` with auth. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:657
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## docs/api/event-hooks.md around lines 110 to 124: the HMAC signature currently

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350581023

- [review_comment] 2025-09-16T03:21:10Z by coderabbitai[bot] (docs/api/event-hooks.md:124)

```text
docs/api/event-hooks.md around lines 110 to 124: the HMAC signature currently
only covers the body which allows replay attacks because the listed
X-Webhook-Timestamp is not bound to the signature; update the docs to require
that the signature is computed over a canonical string that includes the
timestamp (e.g., timestamp + "." + body) and that receivers verify the timestamp
is within a configurable freshness window (e.g., ±N seconds) before accepting
the signature, and document that the server must reject deliveries with
missing/old timestamps or mismatched signatures.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Documented the timestamp-bound HMAC string and freshness window. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Documented the timestamp-bound HMAC string and freshness window. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:696
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/event-hooks.md around lines 246 to 264, the docs currently describe

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350581042

- [review_comment] 2025-09-16T03:21:11Z by coderabbitai[bot] (docs/api/event-hooks.md:264)

```text
In docs/api/event-hooks.md around lines 246 to 264, the docs currently describe
rate limiting and retry policy but do not document idempotency or replay
semantics for DLH replays; add a short subsection stating that replays may
duplicate deliveries and receivers must treat deliveries as potentially
repeated, require an idempotency header (e.g., X-Webhook-Delivery) containing a
unique delivery ID, and optionally include a replay indicator (e.g.,
X-Webhook-Replay: true). Explain receiver behavior: persist the delivery ID with
a configurable TTL, deduplicate by returning a successful 2xx response for
already-processed IDs, use the idempotency key to make non-idempotent operations
safe (skip or noop on duplicate IDs), and document recommended TTL and retention
guidance for the idempotency cache.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Documented replay semantics, idempotency key usage, and the replay header. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Documented replay semantics, idempotency key usage, and the replay header. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:734
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 41-44, the docs show an

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350581052

- [review_comment] 2025-09-16T03:21:11Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:44)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 41-44, the docs show an
incorrect go test command using a filename; replace it with package-aware
commands using -run filters: from repo root use "go test -v ./... -run
'^TestHMACSigner_'" or, when inside the package directory, "cd path/to/package
&& go test -v -run '^TestHMACSigner_'", and update the fenced bash block
accordingly so tests run reliably and with dependencies resolved.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 10 | Updated signature test instructions to use a package-aware `go test -run` pattern. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Updated signature test instructions to use a package-aware `go test -run` pattern. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:775
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 61 to 65 the example test

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350581060

- [review_comment] 2025-09-16T03:21:11Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:65)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 61 to 65 the example test
command uses a filename which is incorrect when run from the repository root;
update the docs to use package paths or a patterned run flag instead: replace
the current `go test -v ./event_filter_test.go` example with a command that runs
tests by package or name, e.g. `go test -v ./... -run '^TestEventFilter_'`, or
alternatively instruct readers to change into the directory containing the test
and run `go test -v` — ensure the doc shows one clear correct command and
removes the filename-based invocation.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 10 | Swapped the event filter docs to use `go test -run` instead of a filename. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Swapped the event filter docs to use `go test -run` instead of a filename. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:811
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 86-90, the docs currently

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350581065

- [review_comment] 2025-09-16T03:21:11Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:90)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 86-90, the docs currently
instruct running a single test file; update the example to run the integration
test by pattern instead of a filename. Replace the existing command with one
that runs the package tests using the -run flag to match the TestWebhookHarness_
tests (e.g., cd test/integration && go test -v -run '^TestWebhookHarness_'), and
ensure the fenced code block language remains bash.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 10 | Integration harness docs now run tests via `-run` within the package. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Integration harness docs now run tests via `-run` within the package. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:849
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 112 to 116, the example

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350581070

- [review_comment] 2025-09-16T03:21:11Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:116)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 112 to 116, the example
command and fenced block are incorrect; replace the existing bash block that
runs `go test -v ./nats_transport_test.go` with a bash fenced block that runs
`cd test/integration && go test -v -run '^TestNATSTransport_'` so the docs use
the correct go test invocation to run the NATS transport tests by name pattern
and ensure the code fence language is "bash".
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 10 | Updated NATS transport instructions to use the package-pattern invocation. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Updated NATS transport instructions to use the package-pattern invocation. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:885
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 136 to 139, the example test

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350581075

- [review_comment] 2025-09-16T03:21:11Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:139)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 136 to 139, the example test
command and fenced block are incorrect for running only DLH tests; replace the
existing bash fenced block (which runs the specific file) with a bash fenced
block that executes the Go test runner with the -run '^TestDLH_' pattern (i.e.,
change the command to: cd test/integration && go test -v -run '^TestDLH_') so
the documentation shows running only DLH tests via the -run flag.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 10 | DLH replay guidance now uses the `-run '^TestDLH_'` pattern. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> DLH replay guidance now uses the `-run '^TestDLH_'` pattern. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:921
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## `

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350581084

- [review_comment] 2025-09-16T03:21:11Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:186)

```text
`
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 182 to 186, the security tests
section currently shows an incorrect go test command that passes a file path;
replace the snippet so it runs the specific test pattern instead (use go test -v
-run '^TestSignatureService_') and ensure the fenced bash block remains intact
and formatted as
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 10 | Security test doc now points to the package pattern instead of a single file. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Security test doc now points to the package pattern instead of a single file. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:957
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 216-224, the docs incorrectly

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350581089

- [review_comment] 2025-09-16T03:21:11Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:224)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 216-224, the docs incorrectly
suggest using file globbing (go test ./*.go) which fails in multi-package repos
and misuses coverage tools; update the commands to run tests across all packages
(replace ./*.go with ./...), use go test -v ./... and go test -v
-coverprofile=coverage.out ./... and then run go tool cover -func=coverage.out
(or -html=coverage.out -o coverage.html) to generate coverage reports so the
instructions work correctly in multi-package projects.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 10 | Coverage instructions now run across packages and include `go tool cover -func`. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Coverage instructions now run across packages and include `go tool cover -func`. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:993
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 228-235, the docs use a

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350581096

- [review_comment] 2025-09-16T03:21:11Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:235)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 228-235, the docs use a
working-directory change and file glob (cd test/integration + go test -v ./*.go)
which doesn't scope tests by package; update the examples to use explicit
package paths instead — replace that block with a single command using the
package path: "go test -v ./test/integration", and similarly update the
"Security Tests Only" example to use the package path (e.g. "go test -v
./test/security_test.go" or "go test -v ./test/integration -run Security"
depending on intended scope) so tests are run by package path rather than
relying on cd and file globs.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 10 | Integration & security examples now rely on explicit package paths. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Integration & security examples now rely on explicit package paths. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:1030
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 239 to 244, the benchmark

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350581101

- [review_comment] 2025-09-16T03:21:12Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:244)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 239 to 244, the benchmark
examples run tests unintentionally and target specific files; update the
commands to filter out tests with -run '^$' and run across packages with ./...
and use a proper benchmark regex for specific benchmarks (e.g.,
-bench='^BenchmarkName$') so benchmarks run only and across packages instead of
executing tests or limiting to ./*.go.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 10 | Benchmark docs now isolate benches with `-run '^$'` and module-wide paths. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Benchmark docs now isolate benches with `-run '^$'` and module-wide paths. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:1069
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In create_postmortem_tasks.py around lines 107 to 108, the dependencies array is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583501

- [review_comment] 2025-09-16T03:22:07Z by coderabbitai[bot] (create_postmortem_tasks.py:108)

```text
In create_postmortem_tasks.py around lines 107 to 108, the dependencies array is
hard-coded with ten POSTMORTEM IDs which is brittle; replace the static list
with a dynamic generation that builds the dependency list from the source of
truth (e.g., the tasks/workers list or a count) so additions/removals stay in
sync — for example, derive task IDs from the tasks collection or generate using
a formatted range like "POSTMORTEM.{:03d}".format(i) and assign that resulting
list to the "dependencies" key.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Coordinator task now derives dependencies from the generated worker tasks instead of a handwritten list. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Coordinator task now derives dependencies from the generated worker tasks instead of a handwritten list. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:1105
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In create_postmortem_tasks.py around lines 114 to 117, the two os.makedirs calls

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583510

- [review_comment] 2025-09-16T03:22:07Z by coderabbitai[bot] (create_postmortem_tasks.py:117)

```text
In create_postmortem_tasks.py around lines 114 to 117, the two os.makedirs calls
use exist_ok=True but lack error handling; wrap each directory creation in a
try/except that catches OSError, detect and handle permission/readonly errors
(e.g. errno.EACCES, errno.EROFS) and any other OS-level failures, and on error
either log a clear message with the path and errno/details and exit with a
non-zero status or re-raise a new exception with contextual information so the
script fails loudly rather than silently.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Added explicit error handling so directory creation fails fast with actionable context. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added explicit error handling so directory creation fails fast with actionable context. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:28
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In demos/lipgloss-transformation.tape lines 1–278, the demo invokes a

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583515

- [review_comment] 2025-09-16T03:22:07Z by coderabbitai[bot] (demos/lipgloss-transformation.tape:278)

```text
In demos/lipgloss-transformation.tape lines 1–278, the demo invokes a
nonexistent ./bin/queue-tui --enhanced; update the tape to run the actual
entrypoint instead (replace the "./bin/queue-tui --enhanced" line with either a
build+run sequence "go build -o bin/queue-tui ./cmd/tui && ./bin/queue-tui" or a
direct run "go run ./cmd/tui" and remove the unsupported "--enhanced" flag), or
alternatively implement and document a matching --enhanced CLI flag in
cmd/tui/main.go that toggles the enhanced view; pick one approach and make the
corresponding change so the demo invocation runs successfully.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Updated the VHS tape to invoke the actual TUI binary and removed the unsupported flag. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Updated the VHS tape to invoke the actual TUI binary and removed the unsupported flag. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:65
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deploy/deploy/data/test.txt lines 1-1 the test data is incorrectly placed

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583522

- [review_comment] 2025-09-16T03:22:07Z by coderabbitai[bot] (deploy/deploy/data/test.txt:1)

```text
In deploy/deploy/data/test.txt lines 1-1 the test data is incorrectly placed
under deploy/deploy; move the file into a Go-style testdata directory such as
testdata/producer/input.ndjson, update any source references, tests, and
CI/deploy manifests to point to the new path, and perform the move as a tracked
rename in git so history is preserved.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Replaced the placeholder text with deterministic NDJSON representing real jobs. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Replaced the placeholder text with deterministic NDJSON representing real jobs. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:138
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deploy/deploy/data/test.txt around lines 1 to 1, the file contains only the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583522

- [review_comment] 2025-09-16T03:22:07Z by coderabbitai[bot] (deploy/deploy/data/test.txt:1)

```text
In deploy/deploy/data/test.txt around lines 1 to 1, the file contains only the
useless line "test file for producer"; replace it with a deterministic NDJSON
fixture representing a real job payload (one JSON object per line) used by the
producer tests. Construct a minimal, valid payload including required fields
(e.g., id, type, payload data, timestamps, and any flags the consumer expects),
ensure values are deterministic (static IDs/timestamps), and save as NDJSON so
each test run consumes identical input for reproducible assertions.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Replaced the placeholder text with deterministic NDJSON representing real jobs. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Replaced the placeholder text with deterministic NDJSON representing real jobs. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:138
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deploy/grafana/dashboards/work-queue.json lines 1-37, the dashboard currently

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583530

- [review_comment] 2025-09-16T03:22:07Z by coderabbitai[bot] (deploy/grafana/dashboards/work-queue.json:37)

```text
In deploy/grafana/dashboards/work-queue.json lines 1-37, the dashboard currently
only defines five panels and lacks operational essentials: add a top-level
"time" range object (e.g., from/to defaults such as now-1h/now) so users can
change window; add a "templating" block with query variables for queue, worker,
and environment (using label_values on relevant metrics) so panels can be
filtered; add an "annotations" block to surface deployments and incidents
(Prometheus expressions such as changes(build_info[...])); convert key panels
(queue length, job failure rate, circuit breaker state) to include alerting
rules with sensible thresholds, evaluation window and for/conditions and add
panel threshold visualization and reduceOptions as needed; and add SLO-focused
panels (error budget burn rate, availability over time, SLO target lines) and an
availability/alerts panel tied to SLO thresholds. Ensure all added fields follow
Grafana dashboard JSON schema and reference the Prometheus datasource/metric
labels used elsewhere in this file.
```

> [!INFO]- **Rejected**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | No | Yes | 6 | The requested dashboard overhaul reaches beyond current metrics coverage and would be speculative right now. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Reviewer note deemed non-actionable. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:175
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/admin-api/Dockerfile around lines 16 to 18, the go build

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583532

- [review_comment] 2025-09-16T03:22:07Z by coderabbitai[bot] (deployments/admin-api/Dockerfile:18)

```text
In deployments/admin-api/Dockerfile around lines 16 to 18, the go build
invocation doesn't strip debug symbols, set a version variable, or trim paths
for reproducible, smaller binaries; change the build step to accept a VERSION
build-arg (ARG VERSION), add -trimpath and -ldflags '-s -w -X
main.version=${VERSION}' to the go build command (keeping CGO_ENABLED=0
GOOS=linux), so the produced admin-api binary is stripped, has embedded version
metadata, and uses reproducible paths.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Docker build now trims paths, strips symbols, and honors a VERSION build arg. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Docker build now trims paths, strips symbols, and honors a VERSION build arg. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:223
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/admin-api/Dockerfile around lines 20 to 23, the image currently

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583546

- [review_comment] 2025-09-16T03:22:08Z by coderabbitai[bot] (deployments/admin-api/Dockerfile:23)

```text
In deployments/admin-api/Dockerfile around lines 20 to 23, the image currently
uses unpinned alpine:latest and runs as root; change to a specific, pinned
Alpine version (e.g., alpine:3.18 or a project-approved tag) and create a
non-root user/group before switching to it: install packages as root, create a
dedicated user and group, create and set ownership of any workdir/home, drop
privileges with USER <username>, and ensure file permissions are set so the
container does not run as root at runtime.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Runtime image now pins Alpine, drops root, and avoids baking configs. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Runtime image now pins Alpine, drops root, and avoids baking configs. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:260
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/admin-api/Dockerfile around lines 26 to 31, the Dockerfile

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583553

- [review_comment] 2025-09-16T03:22:08Z by coderabbitai[bot] (deployments/admin-api/Dockerfile:31)

```text
In deployments/admin-api/Dockerfile around lines 26 to 31, the Dockerfile
currently copies an environment-specific config into the image (COPY
--from=builder /app/configs/admin-api.yaml ./configs/); remove that COPY so
environment-specific configs are not baked into the image and instead rely on
runtime mounting (volume/ConfigMap/Secret). Update the Dockerfile to stop
copying configs, ensure the image expects configs at a runtime path (e.g.,
./configs/) and add a brief comment indicating configs must be mounted at
runtime via your deployment manifests.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 10 | Fixed the table of contents slug so it matches the section heading. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Fixed the table of contents slug so it matches the section heading. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:297
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/docker/Dockerfile.admin-api around lines 20-21, the Go build

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583558

- [review_comment] 2025-09-16T03:22:08Z by coderabbitai[bot] (deployments/docker/Dockerfile.admin-api:21)

```text
In deployments/docker/Dockerfile.admin-api around lines 20-21, the Go build
command should strip debug info, embed a version, and make builds reproducible;
change the build to add -trimpath and -ldflags (e.g. -ldflags "-s -w -buildid=
-X main.version=${VERSION}") to bake in version and remove symbol tables, then
run strip on the resulting binary (or use 'go build' with -ldflags "-s -w" and
follow with 'strip admin-api') so the image is smaller and builds are
reproducible; ensure VERSION is provided via build-arg or environment.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Documented the baseline environment and introduced the new bench payload flag. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Documented the baseline environment and introduced the new bench payload flag. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:335
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/docker/Dockerfile.admin-api around lines 38-40, the COPY line

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583572

- [review_comment] 2025-09-16T03:22:08Z by coderabbitai[bot] (deployments/docker/Dockerfile.admin-api:40)

```text
In deployments/docker/Dockerfile.admin-api around lines 38-40, the COPY line
assumes /app/configs/admin-api.yaml exists in the builder stage which makes the
build fail when it’s absent; fix by guaranteeing the file always exists in the
builder stage (add a step in the builder stage to mkdir -p /app/configs and
create a default admin-api.yaml if missing, or copy a repository default config
into that path) so the final-stage COPY is deterministic and never errors.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Switched the documentation to the interface-based collector example. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Switched the documentation to the interface-based collector example. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:393
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/03_milestones.md around lines 11-13 (and similarly at lines 46-51), the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583575

- [review_comment] 2025-09-16T03:22:08Z by coderabbitai[bot] (docs/03_milestones.md:13)

```text
In docs/03_milestones.md around lines 11-13 (and similarly at lines 46-51), the
table-of-contents uses the incorrect anchor `#gono-go-decision-gates`; update
those links to the correct slug `#go-no-go-decision-gates` (and search the file
for any other occurrences of `gono-go` to replace) so the ToC links point to the
actual section anchor.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Added explicit `BurnRateThresholds` type with documented fields. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added explicit `BurnRateThresholds` type with documented fields. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:492
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/anomaly-radar-slo-budget.md around lines 1–15, add a "Versioning &

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583584

- [review_comment] 2025-09-16T03:22:08Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:15)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 1–15, add a "Versioning &
Deprecation" section that declares supported API versions (e.g., v1), the
compatibility guarantees (minor/patch compatibility, no silent breaking
changes), the breaking-change policy (how breaking changes are evaluated and
approved), the deprecation timeline (minimum 90 days notice before removal), the
changelog/release process (where changes are recorded and how releases are
communicated), and concise migration guidance for clients (examples of typical
migration steps and links to relevant types like SLOConfig, BurnRateThresholds,
AnomalyThresholds, Alert, MetricSnapshot); also add the same section to the
central API docs file (docs/api/_index.md or the repository’s central API docs
entry) so the policy is discoverable project-wide, and ensure any references to
routes (internal/anomaly-radar-slo-budget/handlers.go RegisterRoutes) and types
are linked or cross-referenced for implementer guidance.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | No | No | - | - |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Pending review. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:598
>
> **Alternatives Considered**
> Not documented.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/anomaly-radar-slo-budget.md around lines 36-38 (and also apply the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583587

- [review_comment] 2025-09-16T03:22:09Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:38)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 36-38 (and also apply the
same change at 52-75 and 475-483), the Go import uses the long package path
"github.com/flyingrobots/go-redis-work-queue/internal/anomaly-radar-slo-budget";
update the examples to alias this import to a short, readable identifier (e.g.,
ars or slo) and update all references in the examples accordingly so the code is
concise and consistent across the noted ranges.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Added a public re-export at `pkg/anomaly-radar-slo-budget` and updated docs to reference the stable path. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added a public re-export at `pkg/anomaly-radar-slo-budget` and updated docs to reference the stable path. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:637
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/anomaly-radar-slo-budget.md around lines 36 to 38, the docs instruct

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583587

- [review_comment] 2025-09-16T03:22:09Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:38)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 36 to 38, the docs instruct
users to import an internal package path which is not accessible to consumers;
update the documentation to point to the public exported package path (the
module's published import path for the anomaly-radar-slo-budget package), ensure
the package is exported (move/rename from internal if necessary or add a public
wrapper package), and replace the internal import line with the correct public
import path that consumers can use.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Added a public re-export at `pkg/anomaly-radar-slo-budget` and updated docs to reference the stable path. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added a public re-export at `pkg/anomaly-radar-slo-budget` and updated docs to reference the stable path. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:637
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/anomaly-radar-slo-budget.md around lines 51 to 58, the docs present

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583592

- [review_comment] 2025-09-16T03:22:09Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:58)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 51 to 58, the docs present
two conflicting collector APIs (closures-based SimpleMetricsCollector and an
interface-based QueueMetricsCollector); remove the closures-based
SimpleMetricsCollector snippet and keep the interface-based approach, add an
explicit MetricCollector interface signature description immediately before the
QueueMetricsCollector example so readers see the expected methods and types;
repeat the same cleanup for the other occurrence around lines 446–470 by
deleting the closure example and ensuring the interface signature is documented
prior to the QueueMetricsCollector sample.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | No | No | - | - |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Pending review. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:674
>
> **Alternatives Considered**
> Not documented.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/anomaly-radar-slo-budget.md around lines 81 to 89, the SLOConfig

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583597

- [review_comment] 2025-09-16T03:22:09Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:89)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 81 to 89, the SLOConfig
struct references BurnRateThresholds but that type is not defined; add a new
BurnRateThresholds type immediately after the SLOConfig block with four fields:
FastBurnRate (float64) and FastBurnWindow (time.Duration) for the fast alert
threshold and its evaluation window, and SlowBurnRate (float64) and
SlowBurnWindow (time.Duration) for the slow alert threshold and its evaluation
window, each with brief inline comments explaining units (budget/hour for rates,
time.Duration for windows).
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | No | No | - | - |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Pending review. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:693
>
> **Alternatives Considered**
> Not documented.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/anomaly-radar-slo-budget.md around lines 94-102 (and also apply same

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583603

- [review_comment] 2025-09-16T03:22:09Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:102)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 94-102 (and also apply same
change at lines 193-199), the struct fields lack explicit units and valid
ranges; update the struct comments to include units and ranges (e.g.,
BacklogGrowthWarning/BacklogGrowthCritical: "items/second";
ErrorRateWarning/ErrorRateCritical: "0–1"; LatencyP95Warning/LatencyP95Critical:
"ms"), and add a concise explanatory table or short paragraph immediately
beneath the struct and its JSON examples that lists each field, its unit, and
valid range so readers and JSON consumers have clear expectations.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | No | No | - | - |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Pending review. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:711
>
> **Alternatives Considered**
> Not documented.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/anomaly-radar-slo-budget.md around line 124, there is no

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583606

- [review_comment] 2025-09-16T03:22:09Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:181)

```text
In docs/api/anomaly-radar-slo-budget.md around line 124, there is no
machine-readable API spec; create an OpenAPI 3.1 YAML file that models all
endpoints, request/response schemas, parameters, auth, and example payloads
described in this doc, add it to the repo (e.g., docs/api/openapi.yaml), update
this markdown to link to that file and include a brief note on versioning, and
add a CI job (using OpenAPI Generator CLI or similar) that validates the spec
and generates client SDKs (specify targets, output directory, and caching) on
merge so clients are produced automatically.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/anomaly-radar-slo-budget.md around lines 124-176 (and apply same

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583610

- [review_comment] 2025-09-16T03:22:09Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:176)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 124-176 (and apply same
changes to ranges 179-218, 220-238, 240-270, 272-296, 298-327, 329-350,
352-369): the API endpoint docs lack an auth model, standard error schema, and
explicit Content-Type and status codes; add a short "Authentication &
Authorization" subsection stating the auth scheme (e.g., Bearer token) and
required RBAC roles for the endpoint, add a "Errors" subsection with the
standard JSON error shape (fields: error.code, error.message, error.details,
request_id) and list applicable HTTP error responses (401, 403, 422, 429 where
relevant) and finally specify response Content-Type (application/json) for
request and response examples so every endpoint doc includes auth, error schema,
relevant status codes, and Content-Type.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Added explicit RFC3339/UTC notes under every response example. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added explicit RFC3339/UTC notes under every response example. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:808
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## docs/api/anomaly-radar-slo-budget.md around lines 128 to 176: the response

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583616

- [review_comment] 2025-09-16T03:22:09Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:176)

```text
docs/api/anomaly-radar-slo-budget.md around lines 128 to 176: the response
example lacks an explicit timezone policy for timestamps; add a short sentence
directly beneath each response example block stating that all timestamps are
formatted as RFC3339 in UTC and include the trailing "Z" (e.g., "All timestamps
are RFC3339 in UTC and use the trailing 'Z'"), ensuring the note appears under
every response example in the file.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Added explicit RFC3339/UTC notes under every response example. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added explicit RFC3339/UTC notes under every response example. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:808
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/anomaly-radar-slo-budget.md around lines 132-135 (and also apply

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583619

- [review_comment] 2025-09-16T03:22:09Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:198)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 132-135 (and also apply
same change at 247-249 and 335-337), the "duration" type for query parameters is
not defined; update the docs to state that durations use Go's time.ParseDuration
format and give a short example (e.g., "duration (Go time.ParseDuration format,
e.g., 30m, 1h, 24h, 7h30m)"). Insert this one-line clarification immediately
after each query-parameter list mentioned so callers know the expected format.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/anomaly-radar-slo-budget.md around lines 220 to 238, the PUT

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583622

- [review_comment] 2025-09-16T03:22:09Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:238)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 220 to 238, the PUT
/api/v1/anomaly-radar/config section lacks semantics for validation, unknown
fields, partial updates and immutability; update the doc to state whether PUT
performs a full replace or a deep PATCH-like merge (choose one and describe
behavior for nested objects), enumerate validation rules (e.g.,
availability_target must be between 0 and 1 inclusive, latency_threshold_ms
positive integer, threshold rates between 0 and 1, types), state that invalid
input returns HTTP 422 with a JSON body listing field errors, explain how
unknown fields are handled (reject with 400 or ignore with warning), and call
out any fields that are immutable at runtime and require a restart to change.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Documented PUT as a full replace with explicit validation, 400/422 behaviours, and immutability notes. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Documented PUT as a full replace with explicit validation, 400/422 behaviours, and immutability notes. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:880
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/anomaly-radar-slo-budget.md around lines 246-249, the query param

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583631

- [review_comment] 2025-09-16T03:22:10Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:249)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 246-249, the query param
`max_samples` is underspecified and can lead to huge responses; set and document
a sensible default (e.g., default=1000) and a hard upper bound (e.g.,
max=10_000), add an optional pagination token query param (e.g., `next_cursor`)
and show the paginated response structure including `metrics`, `count`, and
`next_cursor` (opaque token) so callers know how to request subsequent pages;
also mention that `max_samples` cannot exceed the hard limit and that server
will return `count` and `next_cursor=null` when no more data exists.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Documented pagination defaults, limits, and `next_cursor` contract alongside the updated handlers. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Documented pagination defaults, limits, and `next_cursor` contract alongside the updated handlers. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:920
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/anomaly-radar-slo-budget.md around lines 250-266 (and similarly for

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583635

- [review_comment] 2025-09-16T03:22:10Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:266)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 250-266 (and similarly for
lines 339-350), the example metrics response omits p90 while the percentiles
endpoint includes it; update the historical metrics payload to include a
p90_latency_ms field for each metric entry (matching the same format and units
as p50/p95/p99) so both endpoints use the same percentile set, and ensure the
surrounding documentation text reflects that the metrics include
p50/p90/p95/p99.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Added `p90_latency_ms` to the metrics response and updated the descriptive text. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added `p90_latency_ms` to the metrics response and updated the descriptive text. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:958
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/anomaly-radar-slo-budget.md around lines 299 to 327, the response

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583641

- [review_comment] 2025-09-16T03:22:10Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:327)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 299 to 327, the response
fields lack units/definitions; update the documentation by adding explicit unit
notes for budget_utilization, current_burn_rate, and time_to_exhaustion —
specify "budget_utilization: fraction [0,1]", "current_burn_rate: budget/hour
(fraction of total budget consumed per hour)", and "time_to_exhaustion: RFC3339
duration string" under the response example so readers know the domains and
units.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | No | No | - | - |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Pending review. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:995
>
> **Alternatives Considered**
> Not documented.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/anomaly-radar-slo-budget.md around lines 371 to 374, the health

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583645

- [review_comment] 2025-09-16T03:22:10Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:374)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 371 to 374, the health
endpoint HTTP status codes are too limited; add entries for 500, 429, and 206
with brief conditions: 500 Internal Server Error for unexpected/internal
failures, 429 Too Many Requests when collectors or clients are being throttled,
and 206 Partial Content when the endpoint returns partial or degraded data
(include a short parenthetical or one-line condition for each). Ensure
formatting matches the existing bullet list and keep descriptions concise.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Documented additional status codes (500/429/206) for the health endpoint. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Documented additional status codes (500/429/206) for the health endpoint. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:1012
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/anomaly-radar-slo-budget.md around lines 375 to 383, the Start/Stop

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583651

- [review_comment] 2025-09-16T03:22:10Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:383)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 375 to 383, the Start/Stop
section lacks idempotency and authorization details; update the docs to
explicitly define semantics for repeated calls (POST /api/v1/anomaly-radar/start
returns 200 OK with a message "already started" if radar is running, otherwise
202 Accepted or 200 OK when started; POST /api/v1/anomaly-radar/stop returns 200
OK with "already stopped" if not running, otherwise 202/200 when stopping), list
required authentication/authorization (e.g., requires bearer token with role
"slo_admin" or permission "anomaly_radar:manage"), document possible status
codes and example request/response bodies for both success, idempotent-no-op,
and unauthorized (401/403) cases, and add note about concurrency handling
(server guarantees idempotent behavior and returns current state) so callers
know double start/stop are safe.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Documented auth requirements, idempotent responses, and concurrency guarantees for start/stop. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Documented auth requirements, idempotent responses, and concurrency guarantees for start/stop. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:1049
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/anomaly-radar-slo-budget.md around lines 435 to 441, the "Batch

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583660

- [review_comment] 2025-09-16T03:22:10Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:441)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 435 to 441, the "Batch
Operations: Use batch endpoints for efficient data retrieval" bullet references
endpoints that are not documented; remove this bullet or add a new "Batch
endpoints" section detailing the routes and request/response shapes. If
removing, delete bullet 4 and renumber/adjust wording to keep the list coherent;
if adding, create a new subsection immediately after the Performance list that
documents each batch route (path, method), expected request payload, response
schema, and example use-cases so the docs are accurate and not misleading.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Removed the stray batch-operations bullet until matching endpoints exist. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Removed the stray batch-operations bullet until matching endpoints exist. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:1091
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/anomaly-radar-slo-budget.md around lines 488 to 503, the Prometheus

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583663

- [review_comment] 2025-09-16T03:22:10Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:503)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 488 to 503, the Prometheus
exporter incorrectly calls Inc() during scrape which mutates metrics on each
scrape; instead compute counts and set gauges idempotently. Change the loop to
tally active alert counts by severity, call
alertCountVec.WithLabelValues(sev).Set(count) for each severity, and ensure any
previously-exposed severity labels not present are either Set(0) or removed
(e.g., DeleteLabelValues) so the exporter is fully idempotent.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Updated the exporter example to tally counts and use `Set` for idempotent scrapes. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Updated the exporter example to tally counts and use `Set` for idempotent scrapes. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_008.md:1129
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/anomaly-radar-slo-budget.md around lines 519 to 521, the README

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583669

- [review_comment] 2025-09-16T03:22:10Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:521)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 519 to 521, the README
references a /debug endpoint without specifying its contract; either fully
document the endpoint or remove the snippet. If keeping it, add a clear spec:
HTTP method, full path, required auth/headers, request body or query params with
types and validation rules, example request (curl) and example successful and
error responses with status codes and JSON schema; if removing it, delete the
curl snippet and any other mentions of /debug in this doc to avoid misleading
users.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Removed the undocumented `/debug` callout to keep the guide aligned with supported endpoints. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Removed the undocumented `/debug` callout to keep the guide aligned with supported endpoints. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:31
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/chaos-harness.md around lines 259 to 283, the example imports an

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583678

- [review_comment] 2025-09-16T03:22:10Z by coderabbitai[bot] (docs/api/chaos-harness.md:283)

```text
In docs/api/chaos-harness.md around lines 259 to 283, the example imports an
internal package which cannot be imported outside the module; change the import
to the public package path (for example
github.com/flyingrobots/go-redis-work-queue/pkg/chaosharness) and update the
example to use that package name, and if the chaosharness code currently lives
under internal/ move or re-export it under pkg/chaosharness (or otherwise make
it publicly importable) so the example compiles for external users.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Added a `pkg/chaos-harness` re-export and updated the samples to import it via an explicit alias. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added a `pkg/chaos-harness` re-export and updated the samples to import it via an explicit alias. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:70
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/dlq-remediation-ui.md around lines 7 to 10, the documentation

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583687

- [review_comment] 2025-09-16T03:22:11Z by coderabbitai[bot] (docs/api/dlq-remediation-ui.md:10)

```text
In docs/api/dlq-remediation-ui.md around lines 7 to 10, the documentation
currently states the API is unauthenticated; update the docs and implementation
guidance to require authentication and role-based access control for any
endpoints that can requeue or purge jobs. Replace the “no authentication” note
with explicit instructions that anonymous access is banned, list the required
authentication mechanism (e.g., JWT bearer tokens or mTLS) and required RBAC
roles/permissions (e.g., admin:dlq:manage), and add an example request header
and a short note to enforce auth in production deployments and tests. Ensure the
docs also call out auditing/logging for destructive actions and recommend
least-privilege role assignment.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Authentication section now requires bearer tokens/CSRF, documents RBAC scopes, and shows a concrete header example with auditing guidance. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Authentication section now requires bearer tokens/CSRF, documents RBAC scopes, and shows a concrete header example with auditing guidance. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:108
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/dlq-remediation-ui.md around lines 25 to 39, the docs currently

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583693

- [review_comment] 2025-09-16T03:22:11Z by coderabbitai[bot] (docs/api/dlq-remediation-ui.md:39)

```text
In docs/api/dlq-remediation-ui.md around lines 25 to 39, the docs currently
state "max 1000" and mention "rate limits" ambiguously; update the text to
explicitly state that the page_size maximum of 1000 is enforced server-side and
that any rate limits are enforced server-side as well (include where applicable:
page_size and any API rate limiting behavior), e.g., change descriptive cells to
assert server-side enforcement and add a short note clarifying that requests
exceeding page_size or rate limits will be rejected with appropriate HTTP error
codes and messages.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Clarified that `page_size` and rate limits are enforced server-side and documented the rejection behaviour. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Clarified that `page_size` and rate limits are enforced server-side and documented the rejection behaviour. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:149
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/dlq-remediation-ui.md around lines 223-241, the purge-all endpoint

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583698

- [review_comment] 2025-09-16T03:22:11Z by coderabbitai[bot] (docs/api/dlq-remediation-ui.md:241)

```text
In docs/api/dlq-remediation-ui.md around lines 223-241, the purge-all endpoint
currently uses a dangerous confirm=true query parameter; change the docs to
require a POST body with a boolean "confirm" and a "filter" object
(queue,type,error_pattern,start_time,end_time,min_attempts,max_attempts) instead
of query confirm, and document that callers MUST supply an Idempotency-Key
header and a privileged scope/permission to invoke this operation; add
requirements to validate the request body schema, enforce the privileged scope
check on the server, persist the Idempotency-Key with the operation result to
make the purge idempotent and return the stored result (or a 409/appropriate
response) for duplicate keys, and update the endpoint example and params table
to show the request body and required Idempotency-Key header instead of a
confirm query param.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Purge-all now takes a POST body with `confirm`, structured filters, and requires an `Idempotency-Key` header. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Purge-all now takes a POST body with `confirm`, structured filters, and requires an `Idempotency-Key` header. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:188
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/dlq-remediation-ui.md around lines 247 to 257, the JSON example

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583704

- [review_comment] 2025-09-16T03:22:11Z by coderabbitai[bot] (docs/api/dlq-remediation-ui.md:257)

```text
In docs/api/dlq-remediation-ui.md around lines 247 to 257, the JSON example
contains an invalid string "..." inside the successful array; replace it with
valid JSON by either listing only the real example entries (e.g. two sample IDs)
or by truncating the array (e.g. show the first two entries and remove the
ellipsis entirely), keeping the rest of the response fields unchanged so the
example is valid JSON.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 6 | Cleaned the purge-all response example so it is valid JSON without ellipses. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Cleaned the purge-all response example so it is valid JSON without ellipses. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:231
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/dlq-remediation-ui.md around lines 299 to 307, the HTTP status codes

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583715

- [review_comment] 2025-09-16T03:22:11Z by coderabbitai[bot] (docs/api/dlq-remediation-ui.md:307)

```text
In docs/api/dlq-remediation-ui.md around lines 299 to 307, the HTTP status codes
table omits authentication/authorization responses; add rows for 401
Unauthorized and 403 Forbidden with concise descriptions (e.g., "401 -
Unauthorized: Authentication required or invalid credentials" and "403 -
Forbidden: Authenticated but insufficient permissions") and ensure the security
section references these codes where relevant so auth failures are documented
alongside other status codes.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 6 | Added 401/403/429 rows to the status table so auth and throttling failures are documented. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added 401/403/429 rows to the status table so auth and throttling failures are documented. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:268
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 246 to 262, the listed

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2350583735

- [review_comment] 2025-09-16T03:22:12Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:262)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 246 to 262, the listed
performance numbers are presented without context and are misleading; update the
section to replace the single numeric estimates with concrete measurement
details: for each metric state the test environment (hardware, OS, network—e.g.,
localhost vs remote), payload sizes, concurrency level, measurement
tool/version, sample size and duration, and the statistical results (p50/p95/p99
plus mean and standard deviation) and link to raw test scripts/logs; remove or
qualify any numbers that cannot be reproduced and, where appropriate, note
whether the metric was measured under unit, integration, or load-testing
conditions so readers can reproduce the results.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Replaced bare timing bullets with reproducible benchmark tables (environment, workload, p50/p95/p99) and linked commands. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Replaced bare timing bullets with reproducible benchmark tables (environment, workload, p50/p95/p99) and linked commands. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:306
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In .github/workflows/markdownlint.yml around line 3 the reserved YAML key on is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679306

- [review_comment] 2025-09-16T14:19:23Z by coderabbitai[bot] (.github/workflows/markdownlint.yml:3)

```text
In .github/workflows/markdownlint.yml around line 3 the reserved YAML key on is
unquoted and triggers yamllint; either quote the key by changing on: to "on":
(or 'on':) or add a yamllint disable directive for that line (e.g. the
appropriate yamllint disable-line comment) so the linter is silenced while
preserving the existing workflow semantics.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 10 | Quoted the reserved `on` key so yamllint stops complaining. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Quoted the reserved `on` key so yamllint stops complaining. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:347
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## .github/workflows/update-progress.yml lines 16-24: the workflow currently

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679317

- [review_comment] 2025-09-16T14:19:23Z by coderabbitai[bot] (.github/workflows/update-progress.yml:24)

```text
.github/workflows/update-progress.yml lines 16-24: the workflow currently
references actions with mutable tags (actions/checkout@v4 and
actions/setup-python@v5); replace those two uses: entries with the recommended
immutable commit SHAs provided
(actions/checkout@08eba0b27e820071cde6df949e0beb9ba4906955 and
actions/setup-python@a26af69be951a213d495a4c3e4e4022e16d87065) so the workflow
is pinned to fixed commits while preserving existing with: options (fetch-depth
and python-version).
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Pinned checkout/setup-python to immutable SHAs to avoid surprise upgrades. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Pinned checkout/setup-python to immutable SHAs to avoid surprise upgrades. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:383
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In .gitignore around lines 28-33, you need to prevent accidental commits of

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679347

- [review_comment] 2025-09-16T14:19:24Z by coderabbitai[bot] (.gitignore:34)

```text
In .gitignore around lines 28-33, you need to prevent accidental commits of
environment/secret files: add entries to ignore common env filenames (e.g. .env,
.env.* , .env.local, .env.production) while keeping any explicit templates like
.env.example tracked; update the file by appending those patterns (or placing
them before any allow rules) so env files are excluded from git commits.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In append_metadata.py around lines 30 to 33, format_list currently returns "-

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679359

- [review_comment] 2025-09-16T14:19:24Z by coderabbitai[bot] (append_metadata.py:33)

```text
In append_metadata.py around lines 30 to 33, format_list currently returns "-
[]" for empty lists which emits YAML that is a list with a string "[]"; change
format_list to emit a proper YAML empty list (i.e., the key followed by "[]" on
the same line without a dash) when items is empty, and when items exist produce
one "- item" per line with the prefix preserved; then update
generate_yaml_metadata to inline empty collections for soft/enables/provides on
the same line (key: []) rather than emitting a dash line, and apply this same
empty-list handling to soft, enables, and provides.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Empty dependency lists now render as `field: []` while populated lists keep bullet formatting. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Empty dependency lists now render as `field: []` while populated lists keep bullet formatting. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:493
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In append_metadata.py around lines 35 to 59 (and also apply same pattern to

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679374

- [review_comment] 2025-09-16T14:19:24Z by coderabbitai[bot] (append_metadata.py:59)

```text
In append_metadata.py around lines 35 to 59 (and also apply same pattern to
lines 125-131), the script currently mutates files on import; wrap the
procedural code in a main() function and add an if __name__ == "__main__":
main() guard so nothing runs on import, surround file-processing logic with
try/except to log or print errors and continue safely, and ensure the target
DAG/ideas directory exists before writing (os.makedirs(ideas_dir,
exist_ok=True)) to avoid missing-directory errors; update the referenced lines
accordingly to call main and handle exceptions.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | File updates now run under an explicit `main()` with defensive directory creation and error logging. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> File updates now run under an explicit `main()` with defensive directory creation and error logging. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:532
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In BUGS.md around lines 12 to 25, the heartbeat snippet is unsafe: it ignores

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679383

- [review_comment] 2025-09-16T14:19:24Z by coderabbitai[bot] (BUGS.md:25)

```text
In BUGS.md around lines 12 to 25, the heartbeat snippet is unsafe: it ignores
errors, can lose the key TTL, and doesn’t create the heartbeat atomically. Fix
by (1) creating the hb key atomically at start using Redis SET with NX and PX
(or PEXPIRE immediately after a successful SET), (2) renewing using SET with XX
and PX (or PEXPIRE) so you only extend an existing key, (3) adding randomized
jitter to the renewal interval (e.g., renew at TTL/3 ± jitter) to avoid
thundering herds, (4) checking and logging Redis command errors and handling
transient failures with a small retry/backoff or escalate and exit the worker if
renewals repeatedly fail, and (5) ensuring the ticker is stopped and cancel() is
called before performing final LREM/DEL so the background renew goroutine exits
cleanly.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Heartbeat guidance now uses atomic SET NX/XX with jittered renewals and error handling. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Heartbeat guidance now uses atomic SET NX/XX with jittered renewals and error handling. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:571
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In BUGS.md around lines 32-38, the current code scans Redis with pattern

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679399

- [review_comment] 2025-09-16T14:19:25Z by coderabbitai[bot] (BUGS.md:38)

```text
In BUGS.md around lines 32-38, the current code scans Redis with pattern
"jobqueue:*" which floods Redis Cluster across slots; change the reaper to
either (A) maintain and iterate a registry (e.g., a SET of active worker/queue
IDs) and use SMEMBERS to fetch per-worker keys, or (B) use Redis hash-tags so
all jobqueue keys for a worker live in the same slot (e.g.,
jobqueue:{workerID}:processing) and then SCAN only that single slot per worker;
implement one of these approaches and update the description and example in the
doc to show scanning per-worker keys or iterating the SET instead of a global
SCAN "jobqueue:*".
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Guidance now directs operators to maintain a worker registry and hash-tagged processing keys instead of global scans. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Guidance now directs operators to maintain a worker registry and hash-tagged processing keys instead of global scans. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:613
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In BUGS.md around lines 45–53 the scheduler mover uses ZRANGEBYSCORE + ZREM +

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679403

- [review_comment] 2025-09-16T14:19:25Z by coderabbitai[bot] (BUGS.md:53)

```text
In BUGS.md around lines 45–53 the scheduler mover uses ZRANGEBYSCORE + ZREM +
LPUSH in a pipeline which is racy and can double‑enqueue under concurrency;
replace this with an atomic operation: either implement a loop using ZPOPMIN to
pop due members and LPUSH them (ensuring you process only members with score <=
now) or, preferably, write a single Lua script that atomically pops all due
members from the sorted set and pushes them to the list in one server-side
operation; update the code to call the Lua script (or ZPOPMIN loop) and remove
the non‑atomic pipeline approach so deliveries cannot be duplicated.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Replaced the non-atomic pipeline advice with a ZPOPMIN/Lua-based atomic mover. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Replaced the non-atomic pipeline advice with a ZPOPMIN/Lua-based atomic mover. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:653
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In BUGS.md around lines 55–61, the advice to write NDJSON ledger files to local

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679408

- [review_comment] 2025-09-16T14:19:25Z by coderabbitai[bot] (BUGS.md:61)

```text
In BUGS.md around lines 55–61, the advice to write NDJSON ledger files to local
disk risks data loss and PII leakage; change the guidance to require routing
ledger events to a durable sink (S3/Blob store, Kafka, or equivalent) or, if
local append-only files are used, mandate atomic appenders with daily rotation,
size caps, gzip/archive rotation, and documented retention; require payload
redaction/scrubbing of PII before writing (or write only event metadata and
references), implement backpressure and retry/fallback logic on IO failures so
handlers are not blocked, and add documentation for retention/rotation policy
and monitoring/alerts.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Ledger guidance now points to durable sinks (or guarded local appenders) with rotation, retention, and PII scrubbing. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Ledger guidance now points to durable sinks (or guarded local appenders) with rotation, retention, and PII scrubbing. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:692
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In claude_worker.py around lines 29 to 34, only self.my_dir is created but other

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679417

- [review_comment] 2025-09-16T14:19:25Z by coderabbitai[bot] (claude_worker.py:34)

```text
In claude_worker.py around lines 29 to 34, only self.my_dir is created but other
directories (self.finished_dir, self.failed_dir, self.help_dir, and
self.open_tasks_dir) are written to later and may not exist; update the
initialization to mkdir(parents=True, exist_ok=True) for each of those directory
attributes as well (create self.finished_dir, self.failed_dir, self.help_dir,
and self.open_tasks_dir up front) so all required directories exist before any
file operations.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Worker initialisation now ensures open/finished/failed/help directories exist before use. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Worker initialisation now ensures open/finished/failed/help directories exist before use. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:732
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In cmd/job-queue-system/main.go around lines 159-161 (and similarly at 169-171

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679439

- [review_comment] 2025-09-16T14:19:25Z by coderabbitai[bot] (cmd/job-queue-system/main.go:161)

```text
In cmd/job-queue-system/main.go around lines 159-161 (and similarly at 169-171
and 193-201) you’re ignoring json.Marshal errors and printing possibly corrupt
output; replace json.MarshalIndent usage with a json.Encoder writing to
os.Stdout (use encoder.SetIndent("", "  ") for pretty output) and call
encoder.Encode(value), check and handle the returned error (log to stderr and
return/exit with non-zero status or propagate the error) instead of discarding
it; apply the same pattern to every JSON output in runAdmin.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Swapped `json.MarshalIndent` for a shared `json.Encoder` helper that logs encode failures. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Swapped `json.MarshalIndent` for a shared `json.Encoder` helper that logs encode failures. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:770
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In create_postmortem_tasks.py around lines 1-3, the code appends "Z" to a naive

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679449

- [review_comment] 2025-09-16T14:19:25Z by coderabbitai[bot] (create_postmortem_tasks.py:3)

```text
In create_postmortem_tasks.py around lines 1-3, the code appends "Z" to a naive
local datetime which is incorrect; instead import and use a timezone-aware UTC
datetime (e.g., add "from datetime import timezone" and call
datetime.now(timezone.utc)) and emit an ISO8601 UTC timestamp (convert to ISO
format and normalize to Z or use isoformat with UTC) so timestamps are real Zulu
time rather than local time with a trailing "Z".
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Timestamps now use `datetime.now(timezone.utc).isoformat().replace("+00:00", "Z")`. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Timestamps now use `datetime.now(timezone.utc).isoformat().replace("+00:00", "Z")`. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:808
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In create_postmortem_tasks.py around lines 15 to 16, the code appends "Z" to

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679454

- [review_comment] 2025-09-16T14:19:26Z by coderabbitai[bot] (create_postmortem_tasks.py:16)

```text
In create_postmortem_tasks.py around lines 15 to 16, the code appends "Z" to
datetime.now() which produces a fake UTC timestamp; update both timestamp sites
to produce real UTC-aware ISO timestamps by using
datetime.now(timezone.utc).isoformat() (or
datetime.utcnow().replace(tzinfo=timezone.utc).isoformat()), import timezone
from datetime, and remove the manual "Z" concatenation so timestamps are
properly UTC-aware.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Same UTC fix applied to the worker task loop. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Same UTC fix applied to the worker task loop. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:845
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In create_postmortem_tasks.py around lines 27 to 39, the criteria strings still

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679460

- [review_comment] 2025-09-16T14:19:26Z by coderabbitai[bot] (create_postmortem_tasks.py:39)

```text
In create_postmortem_tasks.py around lines 27 to 39, the criteria strings still
contain raw placeholders "{i}" and "{worker}" which won't be interpolated;
replace those literal placeholders with properly formatted/interpolated strings
(e.g., use f-strings or .format with the loop variables i and worker) so the
criteria read the actual worker index and name, and ensure any literal braces
intended to stay are escaped; update the list entries to build their strings
dynamically using the current i/worker variables.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 6 | Criteria now use f-strings, so the rendered instructions include the correct worker ID/name. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Criteria now use f-strings, so the rendered instructions include the correct worker ID/name. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:883
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In create_postmortem_tasks.py around lines 69-70, the timestamp is created with

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679467

- [review_comment] 2025-09-16T14:19:26Z by coderabbitai[bot] (create_postmortem_tasks.py:70)

```text
In create_postmortem_tasks.py around lines 69-70, the timestamp is created with
datetime.now(), which uses local time; change it to produce a UTC timestamp
(e.g., use datetime.utcnow().isoformat() + "Z" or
datetime.now(timezone.utc).isoformat()) and apply the same UTC-fix to the
coordinator task elsewhere in this file so all created_at values are uniform and
in UTC.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Coordinator task timestamps now use the UTC helper as well. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Coordinator task timestamps now use the UTC helper as well. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:921
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In create_review_tasks.py around lines 10-11, the check for "duplicate" is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679474

- [review_comment] 2025-09-16T14:19:26Z by coderabbitai[bot] (create_review_tasks.py:11)

```text
In create_review_tasks.py around lines 10-11, the check for "duplicate" is
case-sensitive so files like "Duplicate..." slip through; change the condition
to test against a lowercased filename (e.g., use 'duplicate' not in f.lower())
while keeping the .json check (you can call f.lower() for the duplicate check
only) so task_id = f[:-5] remains unchanged.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Duplicate filter now lowercases filenames before checking. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Duplicate filter now lowercases filenames before checking. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:958
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In create_review_tasks.py around lines 14 to 21, the code uses a bare except

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679482

- [review_comment] 2025-09-16T14:19:26Z by coderabbitai[bot] (create_review_tasks.py:21)

```text
In create_review_tasks.py around lines 14 to 21, the code uses a bare except
which hides real errors; replace it by catching the specific exceptions that can
occur (e.g., IndexError and ValueError) when splitting/parsing task_id and,
instead of a silent pass, either continue the loop or log the parse failure; for
any truly unexpected exception re-raise or log and raise so real bugs aren’t
masked.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Invalid filenames now log a skip message instead of being silently swallowed. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Invalid filenames now log a skip message instead of being silently swallowed. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:994
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In demos/responsive-tui.tape around lines 72-73 (and also at 129-130, 214-215,

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679493

- [review_comment] 2025-09-16T14:19:26Z by coderabbitai[bot] (demos/responsive-tui.tape:73)

```text
In demos/responsive-tui.tape around lines 72-73 (and also at 129-130, 214-215,
307-308, 365), the script uses "Sleep 3s" which wastes CI minutes; remove these
Sleep commands and instead either remove the pause entirely or replace with a
deterministic check/wait-for-condition (e.g., wait for expected output or
prompt) so the test proceeds immediately when ready; update the surrounding
steps to rely on explicit assertions or readiness checks rather than fixed
sleeps.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Removed all `Sleep` directives from the tape so tests run without artificial waits. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Removed all `Sleep` directives from the tape so tests run without artificial waits. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:1031
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In dependency_analysis.py around lines 23–166, there’s a naming inconsistency:

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679502

- [review_comment] 2025-09-16T14:19:27Z by coderabbitai[bot] (dependency_analysis.py:44)

```text
In dependency_analysis.py around lines 23–166, there’s a naming inconsistency:
keys were defined with hyphens (e.g., "distributed-tracing-integration") but
code still references snake_case names ("distributed_tracing",
"storage_backend"); update all occurrences listed (distributed_tracing at lines
23, 41, 89, 107, 166 and storage_backend at line 112) to use the canonical
hyphenated keys, or alternatively add an explicit alias mapping near the top
that maps "distributed_tracing" -> "distributed-tracing-integration" and
"storage_backend" -> "storage-backends" and then use the canonical names
throughout; after making the edits run the provided rg command to confirm zero
unresolved nodes.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Dependencies now reference `distributed-tracing-integration`/`storage-backends`; alias map handles legacy names. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Dependencies now reference `distributed-tracing-integration`/`storage-backends`; alias map handles legacy names. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:1069
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/admin-api/k8s-redis.yaml around lines 16 to 52, the Pod and

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679524

- [review_comment] 2025-09-16T14:19:27Z by coderabbitai[bot] (deployments/admin-api/k8s-redis.yaml:52)

```text
In deployments/admin-api/k8s-redis.yaml around lines 16 to 52, the Pod and
container lack required security hardening: add automountServiceAccountToken:
false at the pod spec level; add pod-level securityContext with seccompProfile:
{ type: RuntimeDefault }, runAsNonRoot: true and fsGroup: 10001; for the redis
container set imagePullPolicy: Always and pin the image to a digest
(redis:7-alpine@sha256:...); add container securityContext with runAsUser:
10001, runAsGroup: 10001, allowPrivilegeEscalation: false, capabilities: { drop:
["ALL"] } and readOnlyRootFilesystem: true; add an emptyDir mount at /tmp and
keep the redis-data PVC mounted but ensure ownership/permissions allow write by
UID 10001 (use fsGroup and init or securityContext mechanisms); and replace the
readinessProbe exec with a shell command like sh -c 'redis-cli -h 127.0.0.1 -p
6379 ping' for a robust probe.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Added the requested hardening (SA token off, RuntimeDefault, pinned digest, read-only root, /tmp emptyDir, shell probes). |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added the requested hardening (SA token off, RuntimeDefault, pinned digest, read-only root, /tmp emptyDir, shell probes). Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:1147
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/docker/Dockerfile.rbac-token-service around lines 25-27 (and also

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679556

- [review_comment] 2025-09-16T14:19:27Z by coderabbitai[bot] (deployments/docker/Dockerfile.rbac-token-service:27)

```text
In deployments/docker/Dockerfile.rbac-token-service around lines 25-27 (and also
update the similar install at lines 53-54), the image currently installs
ca-certificates, tzdata and wget but the project standard prefers curl for
healthchecks; update the RUN apk --no-cache add ... commands to include curl so
curl is available at runtime for healthcheck scripts and make the same change
for the duplicate install at lines 53-54.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Verified runtime image now ships curl and the healthcheck uses it. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Verified runtime image now ships curl and the healthcheck uses it. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:31
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/00_assessment.md around line 3, the "Last updated: 2025-09-12" header is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679567

- [review_comment] 2025-09-16T14:19:27Z by coderabbitai[bot] (docs/00_assessment.md:3)

```text
In docs/00_assessment.md around line 3, the "Last updated: 2025-09-12" header is
stale for this 2025-09-16 PR; update the timestamp to the current PR date (e.g.,
"Last updated: 2025-09-16") or remove the line entirely if you prefer not to
track last-updated metadata.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Replaced the stale timestamp convention with an evergreen status line per maintainer guidance. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Replaced the stale timestamp convention with an evergreen status line per maintainer guidance. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:68
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/02_release_plan.md around lines 6–7, the release plan text needs

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679574

- [review_comment] 2025-09-16T14:19:28Z by coderabbitai[bot] (docs/02_release_plan.md:7)

```text
In docs/02_release_plan.md around lines 6–7, the release plan text needs
explicit freeze windows, rollout/rollback and go/no‑go gates: add a 48–72h code
freeze before each milestone date (2025-09-26, 2025-10-10, 2025-10-24,
2025-11-07), list the required sign‑off owners for each gate and a short
rollback plan for each release, and add “no High/Critical CVEs” plus
“govulncheck clean” as hard pre‑release gates; update the milestone entries to
include these freeze window dates, the sign‑off owner field, and a one‑line
rollback procedure for each.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Rebuilt the release plan with explicit freezes, owners, rollback notes, and security gates. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Rebuilt the release plan with explicit freezes, owners, rollback notes, and security gates. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:103
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/10_risk_register.md around line 3, the "Last updated" timestamp is stale

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679581

- [review_comment] 2025-09-16T14:19:28Z by coderabbitai[bot] (docs/10_risk_register.md:3)

```text
In docs/10_risk_register.md around line 3, the "Last updated" timestamp is stale
(2025-09-12); update that line to "Last updated: 2025-09-16" and save the file,
and ensure the updated commit includes this change (or implement a quick check
to update the date when editing this file in future).
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Converted the stale timestamp to the new status wording called out by the maintainer. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Converted the stale timestamp to the new status wording called out by the maintainer. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:142
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/admin-api.md around lines 356 to 359, the docs mention a “minimum

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679589

- [review_comment] 2025-09-16T14:19:28Z by coderabbitai[bot] (docs/api/admin-api.md:359)

```text
In docs/api/admin-api.md around lines 356 to 359, the docs mention a “minimum
reason length” but do not define it; update the text to state the exact minimum
(e.g., "Minimum reason length: 10 characters") and the validation behavior: if
the reason is missing or shorter than the minimum the API responds with HTTP 400
and an error body containing code "REASON_REQUIRED" (include a brief example
response schema or one-line note showing HTTP 400 + code REASON_REQUIRED).
Ensure the wording is precise and replaces the vague phrase with the concrete
minimum and the stated HTTP error/code behavior.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Documented the exact minimum reason length and failure envelope in the purge docs. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Documented the exact minimum reason length and failure envelope in the purge docs. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:177
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/anomaly-radar-slo-budget.md around lines 76 to 80, the repo has

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679598

- [review_comment] 2025-09-16T14:19:28Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:119)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 76 to 80, the repo has
mixed go-redis client versions (github.com/go-redis/redis/v8 vs
github.com/redis/go-redis/v9); choose one version (preferably migrate all to v9
or standardize on v8), update all import paths listed in the comment to the
chosen module, update go.mod accordingly, run go mod tidy, run the full test
suite, and fix any API incompatibilities caused by the version change before
merging.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/canary-deployments.md around lines 7 to 11 (and also apply the same

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679606

- [review_comment] 2025-09-16T14:19:28Z by coderabbitai[bot] (docs/api/canary-deployments.md:11)

```text
In docs/api/canary-deployments.md around lines 7 to 11 (and also apply the same
change to lines 753-760), the README currently shows insecure HTTP URLs by
default; change the examples to use HTTPS (and WSS for WebSocket examples) as
the default secure scheme, and move any http/ws examples into a clearly labeled
"Local development" note that indicates using http://localhost only for local
testing; update the Base URL block to use https://... and add a short local-dev
snippet showing the http://localhost URL with an explicit note.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Switched all public examples to HTTPS/WSS and tucked loopback guidance into a local-dev appendix. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Switched all public examples to HTTPS/WSS and tucked loopback guidance into a local-dev appendix. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:254
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/dlq-remediation-pipeline.md around lines 761 to 858, the notify

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679617

- [review_comment] 2025-09-16T14:19:28Z by coderabbitai[bot] (docs/api/dlq-remediation-pipeline.md:927)

```text
In docs/api/dlq-remediation-pipeline.md around lines 761 to 858, the notify
action examples and description omit operational safeguards for outbound egress;
update the docs to define an explicit outbound allowlist for notification
destinations, default and configurable network timeouts, retry policy (attempt
count, backoff), a notification DLQ/escape hatch for failed deliveries, and
clear semantics for partial failures (e.g., per-channel success/failure
reporting, rollback/no-rollback behavior, and how the pipeline marks the job).
Also include recommended defaults and configuration keys, and an example showing
how to configure allowlist, timeouts, retries, and DLQ behavior.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/exactly-once-admin.md around lines 25-33 (and also apply the same

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679626

- [review_comment] 2025-09-16T14:19:29Z by coderabbitai[bot] (docs/api/exactly-once-admin.md:33)

```text
In docs/api/exactly-once-admin.md around lines 25-33 (and also apply the same
change to lines 56-61), the deduplication field "hit_rate" is ambiguous; rename
the field to "hit_percent" and update its value semantics to be a percentage
(e.g., 2.28 means 2.28%), then update the dedup stats JSON example accordingly
and edit the "Fields" documentation block to reflect the new name and explicitly
state that hit_percent is a percentage value (not a fraction) with its units.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Renamed `hit_rate` to `hit_percent` and clarified units everywhere it appears. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Renamed `hit_rate` to `hit_percent` and clarified units everywhere it appears. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:333
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/SLAPS/coordinator-observations.md around lines 114-116 (and also apply

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679641

- [review_comment] 2025-09-16T14:19:29Z by coderabbitai[bot] (docs/SLAPS/coordinator-observations.md:116)

```text
In docs/SLAPS/coordinator-observations.md around lines 114-116 (and also apply
the same fix at 235-242), the text shows “19 tasks completed successfully” while
elsewhere it shows “74 completed,” causing confusion; update the copy to
explicitly annotate that “19” refers to an early snapshot or intermediate
checkpoint and “74” is the final total (or reconcile to a single consistent
number), e.g., add a parenthetical or an extra sentence clarifying the
timeline/source of each number so readers understand they are different
snapshots rather than inconsistent data.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Added context tying the 19-task snapshot to the later 74-task total. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added context tying the 19-task snapshot to the later 74-task total. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:370
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/SLAPS/coordinator-observations.md around lines 121 to 130, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679654

- [review_comment] 2025-09-16T14:19:29Z by coderabbitai[bot] (docs/SLAPS/coordinator-observations.md:130)

```text
In docs/SLAPS/coordinator-observations.md around lines 121 to 130, the
timestamps are missing timezone/offset information; update every timestamp to a
full ISO‑8601 format including date and timezone offset (e.g.
2025-09-16T12:10:00-07:00 or use Z for UTC) so Start Time, End Time and any
other time entries explicitly include timezone/offset; keep human-friendly
labels but ensure machine-parseable ISO strings are used consistently throughout
the file.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Converted the timetable snippet to ISO-8601 with timezone offsets. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Converted the timetable snippet to ISO-8601 with timezone offsets. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:409
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## docs/SLAPS/coordinator-observations.md around lines 249-251: the document

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679656

- [review_comment] 2025-09-16T14:19:29Z by coderabbitai[bot] (docs/SLAPS/coordinator-observations.md:251)

```text
docs/SLAPS/coordinator-observations.md around lines 249-251: the document
contains a profane user quote ("fuck lol thats amazing") which is inappropriate
for a public repo; remove or sanitize the phrase by either replacing it with a
neutral placeholder quote (e.g., "User: 'That's amazing'") or redact the
offending word (e.g., "User: '[redacted] lol that's amazing'"), and ensure the
surrounding sentence still reads naturally and retains the original intent.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Replaced the profane quote with a redacted version to keep the repo family-friendly. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Replaced the profane quote with a redacted version to keep the repo family-friendly. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:447
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/YOU ARE WORKER 6.rb around lines 1 to 5, the file contains non-Ruby

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679659

- [review_comment] 2025-09-16T14:19:29Z by coderabbitai[bot] (docs/YOU ARE WORKER 6.rb:5)

```text
In docs/YOU ARE WORKER 6.rb around lines 1 to 5, the file contains non-Ruby
documentation which breaks RuboCop; either rename the file to a documentation
extension (e.g., docs/YOU ARE WORKER 6.md) or keep the .rb extension but wrap
the entire content in Ruby-safe comment or code fences (e.g., block comments or
a heredoc) so RuboCop ignores it; pick one approach, update the filename or wrap
the content accordingly, and adjust any references in the repo to the new path
if renamed.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Renamed the worker guide to `.md` so lint and tooling stop treating it as Ruby. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Renamed the worker guide to `.md` so lint and tooling stop treating it as Ruby. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:484
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 314 to 320, the examples use

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352679672

- [review_comment] 2025-09-16T14:19:29Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:320)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 314 to 320, the examples use
file globs and improper flag placement; replace them with package-wide
invocations and a proper -run regex: use go test -v ./... -args -debug to enable
verbose logging with the debug arg, and use go test -v ./... -run
'^TestSpecificTest$' to run a single test (anchored regex) so tests target
packages correctly and flags are applied as intended.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Updated the go test snippets to use package-aware invocations with anchored -run. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Updated the go test snippets to use package-aware invocations with anchored -run. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:522
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In AGENTS.md around lines 10 to 41, the table of contents uses Obsidian-style

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683010

- [review_comment] 2025-09-16T14:20:36Z by coderabbitai[bot] (AGENTS.md:41)

```text
In AGENTS.md around lines 10 to 41, the table of contents uses Obsidian-style
wiki links ([[...]]), which won’t render correctly in standard Markdown; convert
each wiki link to a standard Markdown anchor link using the format
[Text](#anchor) where the anchor is the lowercased, hyphenated heading text
(remove extra hashes and punctuation), and update nested entries accordingly so
all TOC links point to the correct section IDs.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Converted the TOC to GitHub-flavoured Markdown anchors. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Converted the TOC to GitHub-flavoured Markdown anchors. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:559
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/admin-api/k8s-deployment.yaml around lines 62 to 116, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683020

- [review_comment] 2025-09-16T14:20:37Z by coderabbitai[bot] (deployments/admin-api/k8s-deployment.yaml:116)

```text
In deployments/admin-api/k8s-deployment.yaml around lines 62 to 116, the
container/pod lacks basic security hardening; add a pod-level and
container-level securityContext to prevent privilege escalation and enforce
least privilege: set spec.template.spec.automountServiceAccountToken: false,
podSecurityContext.runAsNonRoot: true and runAsUser (e.g., 1000) and fsGroup if
needed; in the container securityContext set allowPrivilegeEscalation: false,
readOnlyRootFilesystem: true, runAsNonRoot: true (or match pod runAsUser), and
drop Linux capabilities (capabilities.drop: ["ALL"]); also enable seccomp by
setting securityContext.seccompProfile.type: RuntimeDefault (or annotation for
older k8s) to enforce a default seccomp profile. Ensure values are consistent
between pod and container contexts and do not conflict with existing volume
mounts.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Hardened the admin API deployment with pod/container security contexts and RuntimeDefault seccomp. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Hardened the admin API deployment with pod/container security contexts and RuntimeDefault seccomp. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:596
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/admin-api/k8s-deployment.yaml around lines 117 to 159, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683027

- [review_comment] 2025-09-16T14:20:37Z by coderabbitai[bot] (deployments/admin-api/k8s-deployment.yaml:159)

```text
In deployments/admin-api/k8s-deployment.yaml around lines 117 to 159, the
ingress is fine but you must ensure the Deployment's container probe paths
differ: update the Deployment spec (the container's probes) so
readinessProbe.path is /ready or /readyz and the livenessProbe stays on a
different path (e.g., /health or /live); apply this change in the Deployment
manifest (set readinessProbe.path to /ready or /readyz, keep livenessProbe on
the existing health endpoint) so the readiness check does not use the same
endpoint as liveness.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Ensured readiness probes target `/readyz` while liveness stays on `/healthz`. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Ensured readiness probes target `/readyz` while liveness stays on `/healthz`. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:639
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/docker/docker-compose.yaml around lines 105 to 109 the compose

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683036

- [review_comment] 2025-09-16T14:20:37Z by coderabbitai[bot] (deployments/docker/docker-compose.yaml:109)

```text
In deployments/docker/docker-compose.yaml around lines 105 to 109 the compose
mounts refer to ./grafana/dashboards and ./grafana/datasources but the
repository stores these under deploy/grafana; update the volume paths to point
to ../../deploy/grafana/dashboards and ../../deploy/grafana/datasources (or
alternatively create the missing local ./grafana/datasources directory and add
the required datasource files), ensuring the compose file uses the correct
relative paths to the existing deploy/grafana directories.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Corrected Grafana volume mounts to reference the actual deploy directory. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Corrected Grafana volume mounts to reference the actual deploy directory. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:678
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/docker/rbac-configs/resources.yaml around lines 91-104 (and also

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683042

- [review_comment] 2025-09-16T14:20:37Z by coderabbitai[bot] (deployments/docker/rbac-configs/resources.yaml:104)

```text
In deployments/docker/rbac-configs/resources.yaml around lines 91-104 (and also
check 146-167 and 167-204), the actions catalog is missing the referenced
monitoring/health actions and contains an undefined queues:list alias; add
explicit action entries for "metrics:read" and "health:read" (with description,
risk_level and audit_required) and either add a clear explicit "queues:list"
action entry if you want it aliased to an existing queue read/list permission or
remove any references to queues:list from roles; ensure all roles reference only
actions that are declared in this file.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 6 | Confirmed the monitoring and health actions exist and clarified their descriptions. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Confirmed the monitoring and health actions exist and clarified their descriptions. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:716
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/docker/rbac-configs/token-service.yaml around lines 21 to 24, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683049

- [review_comment] 2025-09-16T14:20:37Z by coderabbitai[bot] (deployments/docker/rbac-configs/token-service.yaml:24)

```text
In deployments/docker/rbac-configs/token-service.yaml around lines 21 to 24, the
JWT config uses a symmetric HS256 key which is unsafe across multiple services;
replace it with an asymmetric algorithm (e.g., RS256 or EdDSA) or PASETO
v4.public, generate a signing private key and a separate public verification key
(use RSA 3072/4096 or Ed25519 as chosen), store the private key only in the
token-service secret and distribute the public key to all verifier services (or
publish via a JWKS endpoint), update the config to reflect algorithm and
key_size appropriate for the chosen algorithm, and add/update deployment docs to
describe the key rotation process (generate new keypair, deploy new public key
to verifiers, switch signer to new private key, and revoke old keys) so
sign/verify roles remain separated.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Switched the token service config to RS256, added key rotation notes, and introduced *_FILE env usage. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Switched the token service config to RS256, added key rotation notes, and introduced *_FILE env usage. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:755
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/kubernetes/admin-api-deployment.yaml around lines 90 to 100, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683057

- [review_comment] 2025-09-16T14:20:37Z by coderabbitai[bot] (deployments/kubernetes/admin-api-deployment.yaml:100)

```text
In deployments/kubernetes/admin-api-deployment.yaml around lines 90 to 100, the
pod-level securityContext exists but the container is missing baseline
hardening; add a container-level securityContext that sets
allowPrivilegeEscalation: false, privileged: false, drops all capabilities
(capabilities.drop: ["ALL"]), sets readOnlyRootFilesystem: true, and configures
a seccompProfile (type: RuntimeDefault) so the container cannot escalate
privileges, has no extra capabilities, uses a read-only root filesystem and
enforces seccomp; keep existing pod runAsNonRoot/runAsUser/fsGroup settings
intact.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Applied the requested security hardening and probe split to the Kubernetes Admin API deployment. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Applied the requested security hardening and probe split to the Kubernetes Admin API deployment. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:801
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/kubernetes/monitoring.yaml around lines 35 to 38, the alert

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683060

- [review_comment] 2025-09-16T14:20:38Z by coderabbitai[bot] (deployments/kubernetes/monitoring.yaml:38)

```text
In deployments/kubernetes/monitoring.yaml around lines 35 to 38, the alert
divides by sum(rate(http_requests_total{app="admin-api"}[5m])) which can be
zero; change the PromQL to guard the denominator (for example wrap the
denominator with clamp_min(..., 1) or otherwise ensure it’s >0 before dividing)
so the expression never performs a division by zero and the rule won’t flap when
traffic is 0.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Wrapped the error-rate denominator with clamp_min to avoid divide-by-zero flapping. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Wrapped the error-rate denominator with clamp_min to avoid divide-by-zero flapping. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:845
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/kubernetes/rbac-monitoring.yaml around lines 45–54, the expr

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683069

- [review_comment] 2025-09-16T14:20:38Z by coderabbitai[bot] (deployments/kubernetes/rbac-monitoring.yaml:54)

```text
In deployments/kubernetes/rbac-monitoring.yaml around lines 45–54, the expr
currently uses raw 5xx RPS instead of a ratio; replace it with a ratio of 5xx
requests to total requests over the same window (and aggregate across labels) —
for example use
sum(rate(http_requests_total{job="rbac-token-service",status=~"5.."}[5m])) /
sum(rate(http_requests_total{job="rbac-token-service"}[5m])) > 0.1 — keeping the
same for/labels/annotations.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Converted the RBAC token service alerts to ratios and proper histogram aggregation. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Converted the RBAC token service alerts to ratios and proper histogram aggregation. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:886
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/kubernetes/rbac-monitoring.yaml around lines 56 to 75, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683075

- [review_comment] 2025-09-16T14:20:38Z by coderabbitai[bot] (deployments/kubernetes/rbac-monitoring.yaml:75)

```text
In deployments/kubernetes/rbac-monitoring.yaml around lines 56 to 75, the
histogram_quantile calls are using raw bucket series instead of aggregated
buckets; replace the inner range vector with sum by (le) over the bucket
streams, e.g. histogram_quantile(0.95, sum by (le)
(rate(http_request_duration_seconds_bucket{job="rbac-token-service"}[5m]))), and
apply the same change for the second alert (the >1.0 and >5.0 thresholds remain
unchanged); ensure both alerts use the sum by (le) aggregation for correct
quantile calculation.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Hardened the RBAC token service deployment with file-based secrets, security contexts, and probe split. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Hardened the RBAC token service deployment with file-based secrets, security contexts, and probe split. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:928
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/kubernetes/rbac-token-service-deployment.yaml around lines

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683084

- [review_comment] 2025-09-16T14:20:38Z by coderabbitai[bot] (deployments/kubernetes/rbac-token-service-deployment.yaml:205)

```text
In deployments/kubernetes/rbac-token-service-deployment.yaml around lines
197–205 (and container block around lines ~231–238) the pod and container
security settings are insufficient: add under pod spec securityContext the
seccompProfile with type: RuntimeDefault and set automountServiceAccountToken:
false; in the rbac-token-service container add a container-level securityContext
with allowPrivilegeEscalation: false, readOnlyRootFilesystem: true, and
capabilities.drop: ["ALL"]; replace the image:
work-queue/rbac-token-service:latest with a pinned immutable image tag (e.g.,
:vX.Y.Z) and change imagePullPolicy to IfNotPresent for pinned images; after
updating, run Checkov/OPA Gatekeeper to verify CKV_K8S_* findings are resolved.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Added persistent forward tracking and cleanup traps to the staging deploy script. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added persistent forward tracking and cleanup traps to the staging deploy script. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:971
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/kubernetes/rbac-token-service-deployment.yaml around lines

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683090

- [review_comment] 2025-09-16T14:20:38Z by coderabbitai[bot] (deployments/kubernetes/rbac-token-service-deployment.yaml:229)

```text
In deployments/kubernetes/rbac-token-service-deployment.yaml around lines
209-229 (and also apply the same change to blocks at 255-262 and 263-271), stop
injecting sensitive secret values (redis-password, rbac-signing-key,
rbac-encryption-key) directly as env values; instead mount the existing Secret
as a volume and update the container spec to volumeMount those secret files,
then change the app to read the secret files from the mounted paths (or if the
app requires env paths, set non-sensitive env vars to the file paths only).
Remove the valueFrom secretKeyRef entries for the secret keys, add a volumes:
entry referencing the Secret name rbac-secrets, add corresponding volumeMounts
with a secure mountPath, and ensure RBAC_SIGNING_KEY, RBAC_ENCRYPTION_KEY and
REDIS_PASSWORD are no longer exposed directly in logs/environment but accessed
from the mounted files; repeat the same adjustments for the other noted blocks.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Replaced the blind sleep with an active poll for the RBAC health check port-forward. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Replaced the blind sleep with an active poll for the RBAC health check port-forward. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:1016
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/scripts/deploy-staging.sh around lines 161-166 (and also apply

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683092

- [review_comment] 2025-09-16T14:20:38Z by coderabbitai[bot] (deployments/scripts/deploy-staging.sh:166)

```text
In deployments/scripts/deploy-staging.sh around lines 161-166 (and also apply
same change to 203-204), the script uses unquoted PID variables and lacks a
cleanup trap; update to quote the PID variables and guard against empty values
(e.g. check [ -n "$PF_PID" ] before calling kill) to avoid globbing/empty-var
issues, and add a trap (e.g. trap cleanup EXIT) plus a small cleanup function
that safely kills quoted PIDs (using kill "$PF_PID" 2>/dev/null || true) to
ensure processes are terminated on exit.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Quoted the monitoring script’s kill command to handle empty or unusual PIDs. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Quoted the monitoring script’s kill command to handle empty or unusual PIDs. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:1063
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/scripts/health-check-rbac.sh around lines 173 to 191, replace the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683098

- [review_comment] 2025-09-16T14:20:39Z by coderabbitai[bot] (deployments/scripts/health-check-rbac.sh:191)

```text
In deployments/scripts/health-check-rbac.sh around lines 173 to 191, replace the
blind sleep used to wait for the port-forward with an active poll that verifies
the local port is accepting TCP connections: implement a loop that checks
localhost:8081 (using a portable method such as bash /dev/tcp/localhost/8081 or
nc if available) with the same $TIMEOUT, exiting success as soon as a connection
can be opened and logging a timeout error if the port never becomes ready; keep
the subsequent HTTP health/metrics checks unchanged and ensure the overall wait
honours the $TIMEOUT variable.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Added idempotency headers, integer duration fields, and dry-run semantics to the batch docs. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added idempotency headers, integer duration fields, and dry-run semantics to the batch docs. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:1105
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/scripts/setup-monitoring.sh around lines 225-226, the kill call

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683107

- [review_comment] 2025-09-16T14:20:39Z by coderabbitai[bot] (deployments/scripts/setup-monitoring.sh:226)

```text
In deployments/scripts/setup-monitoring.sh around lines 225-226, the kill call
uses an unquoted variable which can break if the PID contains spaces or is
empty; change it to quote the variable (use kill "$port_forward_pid" 2>/dev/null
|| true) so the PID is passed safely and to avoid word-splitting or globbing.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | No | No | - | - |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Pending review. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:1148
>
> **Alternatives Considered**
> Not documented.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/dlq-remediation-pipeline.md around lines 149-171, the example

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683113

- [review_comment] 2025-09-16T14:20:39Z by coderabbitai[bot] (docs/api/dlq-remediation-pipeline.md:171)

```text
In docs/api/dlq-remediation-pipeline.md around lines 149-171, the example
response uses a string duration and lacks a clear dry-run and idempotency
contract; change the response to expose duration_ms as an integer (milliseconds)
instead of a string "125ms", explicitly state dry_run is a boolean that
guarantees no state changes when true, and update the POST
/pipeline/process-batch docs to add an Idempotency-Key header (string, optional
but required for at‑least‑once safe retries) and a semantics note that requests
with the same Idempotency-Key must return the original 200 response with an
identical body for 24 hours to prevent duplicate execution.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | No | No | - | - |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Pending review. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:1163
>
> **Alternatives Considered**
> Not documented.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/dlq-remediation-pipeline.md around lines 676 to 687, the documented

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683129

- [review_comment] 2025-09-16T14:20:39Z by coderabbitai[bot] (docs/api/dlq-remediation-pipeline.md:687)

```text
In docs/api/dlq-remediation-pipeline.md around lines 676 to 687, the documented
error response only includes human text and status/timestamp; update the error
model to add a stable machine-readable "code" (string enum) and a "request_id"
(correlation id) field. Modify the example and API schema to include "code" with
a defined enum of values (e.g., rule_not_found, validation_error,
internal_error) and "request_id" as a UUID/string; document that every endpoint
must return a request_id header/value and include it in logs for tracing, and
update any related examples and schema definitions to reflect these new required
fields.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Standardized the DLQ pipeline error envelope with codes plus request IDs. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Standardized the DLQ pipeline error envelope with codes plus request IDs. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_011.md:31
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/dlq-remediation-pipeline.md around lines 701 to 712, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683132

- [review_comment] 2025-09-16T14:20:39Z by coderabbitai[bot] (docs/api/dlq-remediation-pipeline.md:712)

```text
In docs/api/dlq-remediation-pipeline.md around lines 701 to 712, the
rate-limiting section is missing Retry-After semantics, header units, and
scope/precedence of limits; add that 429 responses include a Retry-After header
(with seconds or an HTTP-date) and describe that X-RateLimit-Reset is a Unix
timestamp in seconds (epoch) indicating when the window resets; explicitly state
whether limits are enforced per-principal (API key/token) and/or per-IP and
which takes precedence when both apply (e.g., per-principal overrides per-IP or
both are enforced and the most restrictive applies), and clarify deterministic
behavior for clients (how remaining counters and reset are calculated and which
header to trust for backoff). Ensure the new text is concise and includes header
names, units, 429 use, and scope/precedence rules.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Clarified rate-limit headers, Retry-After semantics, and limit precedence. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Clarified rate-limit headers, Retry-After semantics, and limit precedence. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_011.md:75
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/dlq-remediation-pipeline.md around lines 713 to 731, the docs

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683135

- [review_comment] 2025-09-16T14:20:39Z by coderabbitai[bot] (docs/api/dlq-remediation-pipeline.md:731)

```text
In docs/api/dlq-remediation-pipeline.md around lines 713 to 731, the docs
currently only describe offset pagination which can misbehave under large lists
and concurrent writes; add a cursor-based pagination variant and document stable
sort keys: describe new query params (cursor — opaque token, limit with default
and max), explain that the backend must return a next_cursor (and prev_cursor
optionally) instead of has_next/has_prev, provide example request/response JSON
for the cursor flow, and call out that results must be ordered by one or more
stable sort keys (e.g., created_at ASC + id ASC as tie-breaker) so cursors
remain consistent under concurrent writes; also mention migration guidance
(support both offset and cursor endpoints or parameters) and recommend using
opaque base64 cursors that encode the last item sort-key values.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Documented cursor pagination alongside the legacy offset flow. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Documented cursor pagination alongside the legacy offset flow. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_011.md:121
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/api/dlq-remediation-ui.md around lines 432 to 436, the purge-all example

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683142

- [review_comment] 2025-09-16T14:20:39Z by coderabbitai[bot] (docs/api/dlq-remediation-ui.md:436)

```text
In docs/api/dlq-remediation-ui.md around lines 432 to 436, the purge-all example
uses an unsafe query-string API; replace it with the documented safe POST body
API: change the curl example to a bash code block that sends a POST to
/api/dlq/entries/purge-all with headers Content-Type: application/json,
Authorization: Bearer $TOKEN, and Idempotency-Key: $(uuidgen), and a JSON body
like {"confirm": true, "filter": {"queue":"test-queue"}} (use proper backtick
fencing for the bash block).
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Updated the purge-all example to the safe JSON POST form with idempotency. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Updated the purge-all example to the safe JSON POST form with idempotency. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_011.md:167
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## docs/SLAPS/worker-reflections/claude-008-reflection.md lines 1-16: add a YAML

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683152

- [review_comment] 2025-09-16T14:20:40Z by coderabbitai[bot] (docs/SLAPS/worker-reflections/claude-008-reflection.md:16)

```text
docs/SLAPS/worker-reflections/claude-008-reflection.md lines 1-16: add a YAML
front-matter block at the top containing date and worker_id (e.g., date:
YYYY-MM-DD and worker_id: claude-008) and update the headings to match the other
reflections (use consistent header levels and phrasing such as "SLAPS Reflection
— Worker 8 (claude-008)" and "Summary" / "Tasks" instead of the current informal
titles); ensure the front-matter is the very first content, remove or tighten
any verbose header wording to be parallel with other files, and keep the rest of
the content intact.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Brought the claude-008 reflection in line with the rest of the SLAPS logbook. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Brought the claude-008 reflection in line with the rest of the SLAPS logbook. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_011.md:209
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In README.md lines 44-48 the build section targets Go 1.25, but go.mod (line 3)

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683159

- [review_comment] 2025-09-16T14:20:40Z by coderabbitai[bot] (README.md:55)

```text
In README.md lines 44-48 the build section targets Go 1.25, but go.mod (line 3)
still declares `go 1.24.0`; update go.mod line 3 to `go 1.25` (or `1.25.0`) so
it matches README and the CI workflow files (ci.yml, release.yml,
goreleaser.yml, changelog.yml which use go-version: '1.25.x'), commit the change
and re-run CI to verify everything passes.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## README.md around lines 70 to 86: new users will hit missing Go modules when

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2352683171

- [review_comment] 2025-09-16T14:20:40Z by coderabbitai[bot] (README.md:86)

```text
README.md around lines 70 to 86: new users will hit missing Go modules when
running or building the TUI; add a one‑liner before the run/build examples
telling users to fetch dependencies first by running the Go modules download
command (e.g., "go mod download") so they know to resolve deps prior to go run
or go build.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## .github/workflows/changelog.yml lines 20-29: workflow currently references

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353724315

- [review_comment] 2025-09-16T21:42:53Z by coderabbitai[bot] (.github/workflows/changelog.yml:29)

```text
.github/workflows/changelog.yml lines 20-29: workflow currently references
mutable action tags (actions/checkout@v4 and actions/setup-go@v5); replace those
with the corresponding pinned commit SHAs for each action (e.g.
actions/checkout@<full-commit-sha> and actions/setup-go@<full-commit-sha>) so
the workflow uses immutable references; locate the commit SHAs from the official
GitHub repos for each action and update the uses lines, and do the same for any
other third-party actions in this file to fully eliminate mutable tags.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Confirmed the workflow already wraps the substitution in quotes and preserved fallback behaviour. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Confirmed the workflow already wraps the substitution in quotes and preserved fallback behaviour. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:168
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In .github/workflows/changelog.yml around lines 20 to 24, the checkout step

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353724321

- [review_comment] 2025-09-16T21:42:53Z by coderabbitai[bot] (.github/workflows/changelog.yml:24)

```text
In .github/workflows/changelog.yml around lines 20 to 24, the checkout step
always targets the repository default_branch for tag-triggered runs which can
accidentally write to the default branch; either scope the job to only run on
tag events or make the intent explicit by using the tag ref when the event is a
tag. Fix by adding a workflow-level trigger or job-level condition to only run
on tag events (or distinguish tag vs non-tag runs), and set the checkout ref to
the actual tag ref (or to an explicit variable) when handling a tag so changelog
pushes go to the intended branch only.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## .github/workflows/changelog.yml around lines 35 to 42: the workflow currently

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353724323

- [review_comment] 2025-09-16T21:42:53Z by coderabbitai[bot] (.github/workflows/changelog.yml:42)

```text
.github/workflows/changelog.yml around lines 35 to 42: the workflow currently
sets git user.name/email to generic "github-actions" which can cause
DCO/ownership noise; update the git config to use the official GitHub Actions
bot identity by setting user.name to "github-actions[bot]" and user.email to
"41898282+github-actions[bot]@users.noreply.github.com" before committing so
commits are attributed to the GitHub Actions bot.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## .github/workflows/ci.yml around line 27: CI is using go-version '1.25.x' while

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353724329

- [review_comment] 2025-09-16T21:42:54Z by coderabbitai[bot] (.github/workflows/ci.yml:27)

```text
.github/workflows/ci.yml around line 27: CI is using go-version '1.25.x' while
go.mod declares 'go 1.24.0'; pick one consistent version and update the
corresponding file: either change go.mod to "go 1.25" (and add/verify a
//go:build toolchain directive if your repo uses toolchain management) or change
.github/workflows/ci.yml to use '1.24.x'; after making the change run the full
test suite (and go mod tidy / go vet / go test ./...) on the chosen Go version
before merging.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## .github/workflows/ci.yml around lines 54 to 62: the CI job uses Bash-specific

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353724332

- [review_comment] 2025-09-16T21:42:54Z by coderabbitai[bot] (.github/workflows/ci.yml:62)

```text
.github/workflows/ci.yml around lines 54 to 62: the CI job uses Bash-specific
brace expansion for the loop and lacks timestamps; make the loop POSIX-shell
safe (e.g., use seq or a while counter) so it works under sh/other runners, and
prefix/append each test run with timestamped log lines (use date) to aid
debugging and measure duration; keep set -euo pipefail and ensure any non-zero
test causes workflow failure.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In .github/workflows/goreleaser.yml around lines 13–16, the release job is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353724336

- [review_comment] 2025-09-16T21:42:54Z by coderabbitai[bot] (.github/workflows/goreleaser.yml:16)

```text
In .github/workflows/goreleaser.yml around lines 13–16, the release job is
missing concurrency and pinned action SHAs; add a top-level concurrency block
for the release job (group keyed by the ref or workflow and cancel-in-progress:
true) to serialize tag-triggered runs, and replace each external action version
(e.g., actions/checkout@vX, actions/setup-go@vX, goreleaser-action@vX) with an
explicit commit SHA to pin them to immutable references; ensure every uses:
entry in the job steps points to a specific SHA instead of a floating
major/minor tag.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## .github/workflows/goreleaser.yml around lines 39 to 45: the workflow uses

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353724341

- [review_comment] 2025-09-16T21:42:54Z by coderabbitai[bot] (.github/workflows/goreleaser.yml:45)

```text
.github/workflows/goreleaser.yml around lines 39 to 45: the workflow uses
goreleaser action but doesn’t grant OIDC permission for keyless signing or
provenance/SBOM emission; add repository permissions including "id-token: write"
(and any other required permissions for writing artifacts/provenance if
applicable) in the workflow YAML, and ensure your .goreleaser.yaml has
SBOM/provenance and signing enabled if you intend to produce/sign
provenance/SBOMs; if you aren’t signing/emitting provenance, no change to
permissions is required.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## .github/workflows/markdownlint.yml lines 4-6: the workflow is currently

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353724344

- [review_comment] 2025-09-16T21:42:54Z by coderabbitai[bot] (.github/workflows/markdownlint.yml:6)

```text
.github/workflows/markdownlint.yml lines 4-6: the workflow is currently
triggered on every push and PR; add path filters so the job only runs for
Markdown, workflow, or config changes by adding a paths (or paths-ignore as
preferred) entry under both pull_request and push that includes patterns like
**/*.md, .github/**, and .github/workflows/** (and any other repo config globs
you want to include) so runners are only used when Markdown or workflow/config
files change.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Branch array rewritten as `[main]`, satisfying markdownlint and YAML format guidance. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Branch array rewritten as `[main]`, satisfying markdownlint and YAML format guidance. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:351
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## .github/workflows/markdownlint.yml lines 15-18: the workflow lacks a job timeout

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353724347

- [review_comment] 2025-09-16T21:42:54Z by coderabbitai[bot] (.github/workflows/markdownlint.yml:18)

```text
.github/workflows/markdownlint.yml lines 15-18: the workflow lacks a job timeout
which can cause infinite hangs; add a timeout-minutes setting for the lint job
(e.g., timeout-minutes: 10) under the job definition (right below runs-on or at
the job root) to cap execution time and fail fast if it runs too long.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## .github/workflows/markdownlint.yml around line 17: the workflow uses the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353724349

- [review_comment] 2025-09-16T21:42:54Z by coderabbitai[bot] (.github/workflows/markdownlint.yml:17)

```text
.github/workflows/markdownlint.yml around line 17: the workflow uses the
floating runner "ubuntu-latest" which can change unexpectedly; replace it with a
specific, pinned runner version such as "ubuntu-22.04" (or your project's chosen
LTS like "ubuntu-20.04") by updating the runs-on value to that concrete label so
CI runs are reproducible.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## .github/workflows/markdownlint.yml lines 20-21: the checkout action uses

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353724350

- [review_comment] 2025-09-16T21:42:54Z by coderabbitai[bot] (.github/workflows/markdownlint.yml:21)

```text
.github/workflows/markdownlint.yml lines 20-21: the checkout action uses
fetch-depth: 0 which fetches full git history unnecessarily for a markdown
linter; change the checkout step to use fetch-depth: 1 to speed up the job
(replace 0 with 1 on that line).
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Workflow now ships with least-privilege permissions and concurrency guard as requested. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Workflow now ships with least-privilege permissions and concurrency guard as requested. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:385
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In .goreleaser.yaml around lines 8 to 13, the build configuration lacks

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353792367

- [review_comment] 2025-09-16T22:30:41Z by coderabbitai[bot] (.goreleaser.yaml:13)

```text
In .goreleaser.yaml around lines 8 to 13, the build configuration lacks
reproducibility flags; add the -trimpath flag to the Go ldflags and enable
mod_timestamp (set to a fixed value like 0) in the goreleaser build
configuration so that file paths are trimmed from binaries and timestamps are
stamped consistently across builds; update the ldflags entry to include
-trimpath and add the mod_timestamp setting at the appropriate builds/archives
level.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In .markdownlint.yaml around lines 10 to 13, the MD026 punctuation list includes

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353792378

- [review_comment] 2025-09-16T22:30:41Z by coderabbitai[bot] (.markdownlint.yaml:13)

```text
In .markdownlint.yaml around lines 10 to 13, the MD026 punctuation list includes
".,;:!" but omits "?", so either add the question mark to the allowed
punctuation string or explicitly document that question marks are intentionally
banned; to allow question marks update the punctuation value to include "?"
(e.g., add ? to the string), or if you intend to ban them, add a clarifying
comment above MD026 stating that "?" is intentionally excluded.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## .vscode/extensions.json around lines 2 to 7: the workspace recommendations list

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353792385

- [review_comment] 2025-09-16T22:30:41Z by coderabbitai[bot] (.vscode/extensions.json:7)

```text
.vscode/extensions.json around lines 2 to 7: the workspace recommendations list
lacks the cSpell extension even though a large cSpell dictionary is present in
settings; update the "recommendations" array to include the Spell Checker
extension ID ("streetsidesoftware.code-spell-checker") so new contributors get
prompted to install it (and ensure it is not listed under
"unwantedRecommendations"); add the exact extension ID to the array and save the
file.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## .vscode/extensions.json around line 7: the unwantedRecommendations array is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353792391

- [review_comment] 2025-09-16T22:30:42Z by coderabbitai[bot] (.vscode/extensions.json:7)

```text
.vscode/extensions.json around line 7: the unwantedRecommendations array is
empty but we should proactively block the deprecated PeterJausovec Docker
extension; update the JSON to include "PeterJausovec.vscode-docker" in the
unwantedRecommendations array (preserve other entries, ensure valid JSON syntax,
and avoid duplicates) so VS Code recommends against installing that deprecated
extension.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In .vscode/settings.json around lines 1-8, YAML files aren't configured to

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353792400

- [review_comment] 2025-09-16T22:30:42Z by coderabbitai[bot] (.vscode/settings.json:8)

```text
In .vscode/settings.json around lines 1-8, YAML files aren't configured to
format or fix on save; add a "[yaml]" settings block (or ensure the "yaml"
language id is used) with "editor.formatOnSave": true and, if desired,
"editor.codeActionsOnSave": { "source.fixAll": true } (or other preferred code
action) so YAML manifests/CI are automatically formatted and fixed on save;
optionally reference the YAML extension if your workspace relies on it.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## .vscode/settings.json lines 1-8: trim any trailing whitespace on each line and

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353792400

- [review_comment] 2025-09-16T22:30:42Z by coderabbitai[bot] (.vscode/settings.json:8)

```text
.vscode/settings.json lines 1-8: trim any trailing whitespace on each line and
ensure the file ends with a single newline (final newline present) so diffs are
clean; open the file, remove any trailing spaces, save with UTF-8 and add one
newline character at EOF before committing.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In .vscode/settings.json around line 19, the go.testFlags array currently uses

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353792407

- [review_comment] 2025-09-16T22:30:42Z by coderabbitai[bot] (.vscode/settings.json:19)

```text
In .vscode/settings.json around line 19, the go.testFlags array currently uses
["-race", "-count=1"] but lacks a test timeout, which can allow hung tests to
consume CI time; add a sensible timeout flag (for example "-timeout=2m" or
another project-appropriate duration) to the array so it becomes ["-race",
"-count=1", "-timeout=2m"] to ensure tests fail fast on hangs.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In CHANGELOG.md around lines 9 to 16, replace all placeholder PR tags "[#PR?]"

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353792414

- [review_comment] 2025-09-16T22:30:42Z by coderabbitai[bot] (CHANGELOG.md:16)

```text
In CHANGELOG.md around lines 9 to 16, replace all placeholder PR tags "[#PR?]"
with the real PR number "(#3)" and rewrite the vague "Queue length gauge updater
to surface backlog metrics" line to explicitly name the metric and behavior (for
example: "Periodic updater for the queue_length backlog gauge to surface queue
backlog metrics (#3)"). Make the change for each list item so every bullet ends
with "(#3)" and the queue metric bullet clearly states the metric name and that
a periodic updater/background job updates it to report backlog.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In CHANGELOG.md around line 20, the entry "Smarter rate limiting that sleeps

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353792422

- [review_comment] 2025-09-16T22:30:42Z by coderabbitai[bot] (CHANGELOG.md:20)

```text
In CHANGELOG.md around line 20, the entry "Smarter rate limiting that sleeps
using TTL and jitter for fairness ([#PR?])" is marketing-y and vague; replace it
with a terse, precise description naming the algorithm and behavior such as
"Fixed-window rate limiter with per-key TTL and randomized jitter for backoff
([#PR?])" or the actual algorithm used (e.g., "Token bucket with per-key TTL and
randomized jitter"), keeping it short and factual.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In CHANGELOG.md around line 38, remove the pseudo‑directive line

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353792430

- [review_comment] 2025-09-16T22:30:42Z by coderabbitai[bot] (CHANGELOG.md:38)

```text
In CHANGELOG.md around line 38, remove the pseudo‑directive line
"[request_verification]: Replace placeholder PR numbers with actual references
post-merge." from the user‑facing changelog; delete the entire line (or replace
it with an appropriate finalized reference or normal changelog entry) so
reviewer tags do not appear in public docs.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In create_review_tasks.py around lines 57 and 91, the coverage threshold is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353792440

- [review_comment] 2025-09-16T22:30:42Z by coderabbitai[bot] (create_review_tasks.py:57)

```text
In create_review_tasks.py around lines 57 and 91, the coverage threshold is
inconsistent (one place says "Achieved 90%+ test coverage" while instructions
elsewhere say 80%); choose the correct single requirement (e.g., 90%) and update
both occurrences to the chosen value so the "definition_of_done" and any
instructional text match exactly; ensure any related variable names, comments,
or documentation in the file that reference the coverage percentage are updated
to the same value for consistency.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In demos/responsive-tui.tape around lines 4-6, the script sets Theme "Tokyo

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353792445

- [review_comment] 2025-09-16T22:30:43Z by coderabbitai[bot] (demos/responsive-tui.tape:6)

```text
In demos/responsive-tui.tape around lines 4-6, the script sets Theme "Tokyo
Night" and FontFamily "Fira Code" which are non-deterministic across runners;
update to either remove these assumptions or add deterministic fallbacks: either
drop the Theme/FontFamily lines, or change FontFamily to a comma-separated
fallback (e.g., "Fira Code, monospace") and ensure the test
environment/container includes the Fira Code font (or bundle the font into the
test image) and pin the theme resource so rendering is reproducible across
runners.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In demos/responsive-tui.tape around line 9, the TypingSpeed is set to 80ms which

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353792455

- [review_comment] 2025-09-16T22:30:43Z by coderabbitai[bot] (demos/responsive-tui.tape:9)

```text
In demos/responsive-tui.tape around line 9, the TypingSpeed is set to 80ms which
overly slows the demo; lower it to a more reasonable value (e.g., 10–25ms) to
reduce runtime. Edit the tape file to change "Set TypingSpeed 80ms" to a faster
value (pick one consistent with other demos) so the demo runs snappily while
preserving readability.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In demos/responsive-tui.tape around line 10, stop hard-coding zsh; change the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353792460

- [review_comment] 2025-09-16T22:30:43Z by coderabbitai[bot] (demos/responsive-tui.tape:10)

```text
In demos/responsive-tui.tape around line 10, stop hard-coding zsh; change the
"Set Shell \"zsh\"" directive to a portable shell (e.g., "Set Shell \"bash\"")
or remove the directive to use the system default shell so CI images without zsh
won't fail; update the line to use bash and ensure any script syntax in the tape
is compatible with bash.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In demos/responsive-tui.tape around lines 12 to 16, the demo relies on the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353792463

- [review_comment] 2025-09-16T22:30:43Z by coderabbitai[bot] (demos/responsive-tui.tape:16)

```text
In demos/responsive-tui.tape around lines 12 to 16, the demo relies on the
host's locale so emoji and box-drawing characters can render incorrectly;
explicitly set the UTF-8 locale at the top of the tape or before printing UI
content (for example export LANG and LC_ALL to an en_US.UTF-8 or similar UTF-8
locale, or invoke a locale-safe wrapper) and add a short runtime check/fallback
that warns and exits if UTF-8 is not available so the emojis/box drawing render
consistently.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In demos/responsive-tui.tape around lines 28-29 (also apply the same fix at

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353792466

- [review_comment] 2025-09-16T22:30:43Z by coderabbitai[bot] (demos/responsive-tui.tape:28)

```text
In demos/responsive-tui.tape around lines 28-29 (also apply the same fix at
85-86, 144-145, 231-232, 373-374): the script sets a fake terminal width via
"export COLUMNS=35" but never restores or unsets it, leaking the environment
variable to the rest of the session; change each snippet to save the prior
COLUMNS (e.g., OLD_COLUMNS="$COLUMNS"), set the test value, then after the test
restore the prior value (if non-empty) or unset COLUMNS (e.g., if [ -z
"$OLD_COLUMNS" ]; then unset COLUMNS; else export COLUMNS="$OLD_COLUMNS"; fi) so
downstream commands do not inherit the fake width.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In demos/responsive-tui.tape around lines 30 to 73, the script emits an

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353792467

- [review_comment] 2025-09-16T22:30:43Z by coderabbitai[bot] (demos/responsive-tui.tape:72)

```text
In demos/responsive-tui.tape around lines 30 to 73, the script emits an
excessive sequence of individual "Type"/"Enter" steps to produce a static block;
compress these into a single paste/heredoc operation (e.g., one cat << 'EOF' ...
EOF paste) so the entire block is inserted in one step. Replace the repeated
Type/Enter lines with a single paste action that contains the full ASCII UI,
ensure correct quoting/escaping so no extra interpolation occurs, and remove the
redundant keystroke steps so the tape uses a single bulk-paste operation
supported by VHS.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In dependency_analysis.py around lines 213-218, the entry lists "json_editor" as

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353792470

- [review_comment] 2025-09-16T22:30:43Z by coderabbitai[bot] (dependency_analysis.py:218)

```text
In dependency_analysis.py around lines 213-218, the entry lists "json_editor" as
both a hard dependency and a provided capability; update the provides list to
disambiguate by renaming the provided capability to "json_editor_ui" (leave the
hard dependency as "json_editor") and search/replace any local references to the
old provided name so consumers use "json_editor_ui".
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In auto_commit.sh around lines 5 to 16, there is no preflight check for git

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814673

- [review_comment] 2025-09-16T22:46:55Z by coderabbitai[bot] (auto_commit.sh:16)

```text
In auto_commit.sh around lines 5 to 16, there is no preflight check for git
configuration so the loop will repeatedly fail if git user.name or user.email
are not set; add a startup preflight function that verifies git is available and
inside a git repo (git rev-parse --is-inside-work-tree), then checks git config
--get user.name and git config --get user.email and exits immediately with a
non-zero status and an explanatory stderr message if any check fails; call this
preflight function once at script startup before entering the main loop so the
script fails fast instead of churning on commit errors.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In auto_commit.sh around lines 51 to 61, the commit message is built via a

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814675

- [review_comment] 2025-09-16T22:46:55Z by coderabbitai[bot] (auto_commit.sh:61)

```text
In auto_commit.sh around lines 51 to 61, the commit message is built via a
subshell heredoc which can introduce trailing-newline quirks; replace that with
two explicit strings and pass them to git commit using two -m flags: build a
subject variable like "chore(slaps): auto-sync progress - $DONE done / $OPEN
open" and a body variable containing the Stats block without relying on
command-substitution heredoc, then run git commit -m "$subject" -m "$body" and
preserve the same multiline body formatting within the body string.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In auto_commit.sh around lines 62 to 75, the script always pushes to origin and

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814676

- [review_comment] 2025-09-16T22:46:55Z by coderabbitai[bot] (auto_commit.sh:75)

```text
In auto_commit.sh around lines 62 to 75, the script always pushes to origin and
uses an unescaped @{u} which triggers SC1083; change it to detect the actual
upstream remote/branch using an escaped ref (e.g. capture upstream_ref=$(git
rev-parse --abbrev-ref --symbolic-full-name '\@{u}' 2>/dev/null)), if
upstream_ref is non-empty split it into upstream_remote and upstream_branch and
push to that remote/branch (git push "$upstream_remote"
"$current_branch:$upstream_branch"), otherwise fall back to creating an upstream
with --set-upstream (e.g. git push --set-upstream origin "$current_branch"), and
keep the existing success/failure logging.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deploy/data/test.txt (lines 1-1) this stray test artifact should not be

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814679

- [review_comment] 2025-09-16T22:46:55Z by coderabbitai[bot] (deploy/data/test.txt:1)

```text
In deploy/data/test.txt (lines 1-1) this stray test artifact should not be
shipped; either delete the file from deploy/ or relocate it to a proper test
fixture path such as producer/testdata/input.txt (preferred for Go tooling), and
if you relocate it add a short README.md next to it explaining its purpose and
format so CI/images don’t pick up deploy/ artifacts by accident.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deploy/data/test.txt lines 1-1, this test file must never be included in

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814679

- [review_comment] 2025-09-16T22:46:55Z by coderabbitai[bot] (deploy/data/test.txt:1)

```text
In deploy/data/test.txt lines 1-1, this test file must never be included in
built images or releases; add entries to project ignore/config files so it’s
always excluded. Update .dockerignore and .helmignore to include
deploy/data/test.txt (and deploy/data/ as appropriate), and if you use
goreleaser, add exclusion patterns for test/ or deploy/data/test.txt (or fixture
paths) under the archives/exclude section in .goreleaser.yaml so packaging never
includes this file.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deploy/grafana/dashboards/work-queue.json around lines 10 to 12, the PromQL

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814681

- [review_comment] 2025-09-16T22:46:55Z by coderabbitai[bot] (deploy/grafana/dashboards/work-queue.json:12)

```text
In deploy/grafana/dashboards/work-queue.json around lines 10 to 12, the PromQL
currently computes a global p95 across all queues; change the query to aggregate
histograms by both le and queue (sum by (le, queue) (rate(...))) so
histogram_quantile(0.95, ...) is evaluated per-queue, and set the panel/metric
legendFormat to include the queue label (e.g. {{queue}}) so operators can
identify which queue the p95 belongs to.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## deploy/grafana/dashboards/work-queue.json around lines 65-86: the Stat panel

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814685

- [review_comment] 2025-09-16T22:46:55Z by coderabbitai[bot] (deploy/grafana/dashboards/work-queue.json:86)

```text
deploy/grafana/dashboards/work-queue.json around lines 65-86: the Stat panel
currently computes a single value across multiple series unpredictably (it picks
a sample); change the reduceOptions to aggregate across series by using the
"sum" calculation (set reduceOptions.calcs to ["sum"]) so the panel shows total
active workers, and optionally add a value text override "{{__value.raw}}
workers" and explicit thresholds with 0 -> red and >0 -> green.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/admin-api/k8s-redis.yaml around lines 7 to 12, the ServiceAccount

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814690

- [review_comment] 2025-09-16T22:46:56Z by coderabbitai[bot] (deployments/admin-api/k8s-redis.yaml:12)

```text
In deployments/admin-api/k8s-redis.yaml around lines 7 to 12, the ServiceAccount
is created without disabling token automounting; add
automountServiceAccountToken: false to the ServiceAccount spec (i.e., under
metadata add the automountServiceAccountToken field at the same indentation
level) so pods using this SA do not automatically get a token mounted.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/admin-api/k8s-redis.yaml around lines 31-34 (and also lines

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814692

- [review_comment] 2025-09-16T22:46:56Z by coderabbitai[bot] (deployments/admin-api/k8s-redis.yaml:34)

```text
In deployments/admin-api/k8s-redis.yaml around lines 31-34 (and also lines
49-51), the securityContext uses runAsNonRoot: true with fsGroup: 1000 which
collides with common host UIDs; choose a high, non‑host UID/GID (e.g. >= 100000
or from your cluster's allocated range) and set runAsUser and runAsGroup to that
UID/GID and update fsGroup to the same high GID; ensure the chosen ID is used
consistently for both container specs so the pod runs as non‑root without
conflicting with host users.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/admin-api/k8s-redis.yaml around lines 40 to 45, the Redis command

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814693

- [review_comment] 2025-09-16T22:46:56Z by coderabbitai[bot] (deployments/admin-api/k8s-redis.yaml:49)

```text
In deployments/admin-api/k8s-redis.yaml around lines 40 to 45, the Redis command
is missing an explicit data directory; add the flag --dir /data to the command
array so Redis writes to /data, and ensure the Pod spec includes a volumeMount
for /data backed by a persistent volume (or emptyDir if ephemeral) and a
corresponding volume or PVC entry in the deployment.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/admin-api/k8s-redis.yaml around lines 66 to 79, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814695

- [review_comment] 2025-09-16T22:46:56Z by coderabbitai[bot] (deployments/admin-api/k8s-redis.yaml:79)

```text
In deployments/admin-api/k8s-redis.yaml around lines 66 to 79, the
liveness/readiness probes use exec redis-cli ping; replace these with tcpSocket
probes against the Redis port (typically containerPort 6379) to avoid relying on
an external binary. For both probes remove the exec block and add tcpSocket:
port: 6379, preserving or adjusting initialDelaySeconds and periodSeconds as
appropriate; ensure probe entries remain under the container spec and validate
YAML indentation.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/01_product_roadmap.md around lines 54 to 56, the dependency references

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814702

- [review_comment] 2025-09-16T22:46:56Z by coderabbitai[bot] (docs/01_product_roadmap.md:56)

```text
In docs/01_product_roadmap.md around lines 54 to 56, the dependency references
use placeholder PR numbers (#123, #145, #130); replace each stub PR reference
with the actual PR number or a direct link to the corresponding doc/issue before
GA (or change to the appropriate document link anchor), and ensure owner names
remain intact; update the three lines so each dependency points to a real,
resolvable reference (PR URL or docs link) or remove the PR token if no final
reference exists.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/13_release_versioning.md around lines 17–19, the changelog rules are

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814707

- [review_comment] 2025-09-16T22:46:56Z by coderabbitai[bot] (docs/13_release_versioning.md:19)

```text
In docs/13_release_versioning.md around lines 17–19, the changelog rules are
incomplete: update this section to explicitly require Conventional Commits
v1.0.0 semantics (type(scope)!: description / type(scope): description), mandate
allowed types (feat, fix, docs, ci, refactor, test, chore, perf, build), require
optional scope, optional body, and FOOTER/BREAKING CHANGE format for breaking
changes; add mapping rules for generating standardized CHANGELOG.md sections
(Features, Fixes, Docs, CI, Refactor, Tests, Chore, Performance, Build) and
rules for incrementing semver based on types/BREAKING CHANGE, include a short
example commit and breaking-change example, and add enforcement notes to run
commitlint and CI hook to validate commits and produce machine‑readable
changelog output for release tooling.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/13_release_versioning.md around line 23, the phrase "Ensure CI green;

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814709

- [review_comment] 2025-09-16T22:46:57Z by coderabbitai[bot] (docs/13_release_versioning.md:23)

```text
In docs/13_release_versioning.md around line 23, the phrase "Ensure CI green;
govulncheck passes; tests (unit/race/e2e) pass." is vague; replace it with an
explicit list of CI gates to enforce (e.g., status: build passes, lint/format
checks, govulncheck report no findings, unit tests, race-detector tests,
integration/e2e tests, and required approvals). Update the line to enumerate
those checks and any minimum thresholds (e.g., coverage or no vuln findings) and
reference the exact CI job names used in the pipeline so reviewers know which
gates must be green.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/13_release_versioning.md around line 26, the doc currently says to

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814712

- [review_comment] 2025-09-16T22:46:57Z by coderabbitai[bot] (docs/13_release_versioning.md:26)

```text
In docs/13_release_versioning.md around line 26, the doc currently says to
"ensure" supply-chain artifacts but lacks concrete, blocking verification steps;
update the section to include explicit, copy-paste verification commands for (1)
cosign container signature verification bound to the tag and OIDC issuer, (2)
slsa-verifier provenance verification against the release.intoto.jsonl and the
repo+tag, and (3) SBOM emission via syft producing spdx-json, and instruct users
to replace placeholders (org/repo, TAG, registry/image@digest, provenance path,
artifacts) and to run these commands in a failing/CI-blocking mode (e.g., run in
a shell with errexit or check exit codes) so any verification failure causes the
release pipeline to stop.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/13_release_versioning.md around lines 27 to 32, the instructions

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814715

- [review_comment] 2025-09-16T22:46:57Z by coderabbitai[bot] (docs/13_release_versioning.md:32)

```text
In docs/13_release_versioning.md around lines 27 to 32, the instructions
currently create annotated tags and push them but should prefer signed tags,
avoid lightweight tags, and ensure exactly one ref is pushed; update the example
to use git tag -s for signed tags (with a descriptive "release: vX.Y.Z[-pre]"
message), add a commented annotated-tag fallback for CI environments that cannot
sign, and ensure git push references the exact tag name (push one ref) so the
documentation shows using signed tags by default and the annotated fallback as a
comment.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/SLAPS/worker-reflections/claude-006-reflection.md around lines 8 to 12,

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814720

- [review_comment] 2025-09-16T22:46:57Z by coderabbitai[bot] (docs/SLAPS/worker-reflections/claude-006-reflection.md:12)

```text
In docs/SLAPS/worker-reflections/claude-006-reflection.md around lines 8 to 12,
the file duplicates the date in the body which may drift; remove the "Date:
September 14, 2025" line (or alternatively remove the front-matter date) so
there is a single canonical source of truth, and update the file to render the
date from front matter (or keep only the body date) and delete the redundant
body line.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/SLAPS/worker-reflections/claude-006-reflection.md around line 12, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814723

- [review_comment] 2025-09-16T22:46:57Z by coderabbitai[bot] (docs/SLAPS/worker-reflections/claude-006-reflection.md:12)

```text
In docs/SLAPS/worker-reflections/claude-006-reflection.md around line 12, the
"SLAPS Experiment Duration: [Session duration]" placeholder must be replaced
with the real session duration or removed entirely; update the line to either
"SLAPS Experiment Duration: X minutes" (use the accurate duration) or delete the
whole line and any trailing empty line, then save and commit with a message like
"docs: fill/remove SLAPS experiment duration placeholder".
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/SLAPS/worker-reflections/claude-006-reflection.md around lines 20-23

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814724

- [review_comment] 2025-09-16T22:46:57Z by coderabbitai[bot] (docs/SLAPS/worker-reflections/claude-006-reflection.md:23)

```text
In docs/SLAPS/worker-reflections/claude-006-reflection.md around lines 20-23
(and similarly 73-76), replace the vague descriptions with concrete symbols and
versions: name the exact miniredis function signatures and module version you
hit (e.g., miniredis/v2 redis.Set(key, val) vs redis.SetEx with TTL in v2.32.0),
and fully qualify struct types/fields (e.g., pkg.ClusterConfig.Environment,
pkg.ClusterConfig.Region) including commit hashes or git refs where the shape
differed (commit abc123). Update sentences to show the exact function/field
names, versions, and a short code-like example of expected vs actual API so
readers can reproduce the mismatch.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## docs/SLAPS/worker-reflections/claude-006-reflection.md around line 117: there is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814725

- [review_comment] 2025-09-16T22:46:57Z by coderabbitai[bot] (docs/SLAPS/worker-reflections/claude-006-reflection.md:117)

```text
docs/SLAPS/worker-reflections/claude-006-reflection.md around line 117: there is
a stray internal '---' separator that can be mis-parsed as YAML front‑matter;
replace that internal '---' with '***' (or an explicit <hr/> or remove it) so
only the file header remains as YAML front matter, save and commit the change;
optionally grep the docs/SLAPS tree for other files containing internal '---'
separators and apply the same replacement to avoid front‑matter parsing issues.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/YOU ARE WORKER 6.rb around lines 3 to 4, the sentence "You are a worker

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814728

- [review_comment] 2025-09-16T22:46:57Z by coderabbitai[bot] (docs/YOU ARE WORKER 6.rb:4)

```text
In docs/YOU ARE WORKER 6.rb around lines 3 to 4, the sentence "You are a worker
in the SLAPS task execution system. Your job is to claim and execute tasks for
the go-redis-work-queue project." was split across two lines with an extra
indent; fix it by joining into a single line without the stray indentation so
the sentence is one continuous line.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Role description now uses vendor-neutral wording and adds safety guidance. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Role description now uses vendor-neutral wording and adds safety guidance. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:251
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/YOU ARE WORKER 6.rb around line 28, the mv command may mis-handle

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353814732

- [review_comment] 2025-09-16T22:46:58Z by coderabbitai[bot] (docs/YOU ARE WORKER 6.rb:28)

```text
In docs/YOU ARE WORKER 6.rb around line 28, the mv command may mis-handle
filenames that begin with a dash; update the command to include the
end-of-options marker so it becomes mv -n --
"slaps-coordination/open-tasks/P1.T001.json" "slaps-coordination/claude-001/" to
ensure paths starting with “-” are treated as operands rather than options.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In append_metadata.py around lines 57 to 113, remove the hard-coded

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856153

- [review_comment] 2025-09-16T23:20:20Z by coderabbitai[bot] (append_metadata.py:138)

```text
In append_metadata.py around lines 57 to 113, remove the hard-coded
infrastructure_nodes and instead import the canonical infrastructure list and
normalization helpers from dependency_analysis; build node_map keyed by the
normalized name (e.g., use normalize_name(name)) while storing the original
display name in the node dict, normalize every dependency name before doing
lookups so feature→feature edges are not dropped due to kebab/snake differences,
and replace any manual path concatenation with os.path.join(ideas_dir, ...) when
resolving spec paths.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In claude_worker.py lines 1-5, the module currently has a short descriptive

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856159

- [review_comment] 2025-09-16T23:20:21Z by coderabbitai[bot] (claude_worker.py:5)

```text
In claude_worker.py lines 1-5, the module currently has a short descriptive
comment but lacks a proper module docstring describing usage and contracts;
replace the placeholder with a real triple-quoted module docstring that
documents how to run the worker, required environment variables/CLI args, the
expected coordination directory layout, the exact JSON schema for task files
(fields, types, required/optional), file naming and lock/claim semantics,
error-handling expectations and return codes, and examples of typical input and
output; keep it concise, accurate, and in reStructuredText or Google-style so
callers and future maintainers can implement and validate producers/consumers
against these contracts.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In claude_worker.py around lines 7 to 16, replace ad-hoc prints with a proper

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856164

- [review_comment] 2025-09-16T23:20:21Z by coderabbitai[bot] (claude_worker.py:16)

```text
In claude_worker.py around lines 7 to 16, replace ad-hoc prints with a proper
Python logger: import the logging module, configure basic logging (level and
format) or load config, and create a module-level logger via
logging.getLogger(__name__); then replace all print(...) calls across the file
with appropriate logger methods (logger.debug/info/warning/error/critical)
according to message severity. Ensure logger configuration happens once at
process startup (not inside functions) and avoid printing sensitive data; keep
fallback to stdout only for local dev if needed.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In claude_worker.py around lines 121-123, the unknown-status branch currently

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856166

- [review_comment] 2025-09-16T23:20:21Z by coderabbitai[bot] (claude_worker.py:130)

```text
In claude_worker.py around lines 121-123, the unknown-status branch currently
only logs and returns, leaving the task file orphaned in my_dir; modify this
branch to atomically move the task's file from my_dir into the help queue
directory (e.g., help_dir or my_dir/help) and write or attach contextual
metadata (task_id, status value, timestamp, and any error/trace info) so humans
can triage; handle and log any filesystem errors and ensure the function returns
a value indicating the task was requeued to help.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In claude_worker.py around lines 124-127, the except block for

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856170

- [review_comment] 2025-09-16T23:20:21Z by coderabbitai[bot] (claude_worker.py:134)

```text
In claude_worker.py around lines 124-127, the except block for
json.JSONDecodeError and OSError currently just prints and returns False leaving
the task file in place; instead, create (if missing) a failed-tasks directory
next to my_dir and atomically move or write a failure payload there that records
the original task file name, the error message/stack, timestamp, and
(optionally) the original file contents; ensure you capture the exception as
err, build a JSON payload with those fields, write it to failed-tasks using a
deterministic filename (e.g. originalname + ".failed.json" or a UUID), remove or
rename the original task file so it is no longer left in my_dir, and then return
False after the move/write completes.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In claude_worker.py around lines 163 to 167 the _persist_task function writes

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856174

- [review_comment] 2025-09-16T23:20:21Z by coderabbitai[bot] (claude_worker.py:174)

```text
In claude_worker.py around lines 163 to 167 the _persist_task function writes
JSON to disk but does not flush and fsync, risking loss on crash; modify the
function to open the file, write the JSON, call handle.flush() and
os.fsync(handle.fileno()) before closing, ensure parent directories are created
as before, and keep encoding="utf-8"; also handle exceptions if desired and
avoid changing the function signature.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In claude_worker.py around lines 171 to 180, the argument range check is done

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856177

- [review_comment] 2025-09-16T23:20:21Z by coderabbitai[bot] (claude_worker.py:187)

```text
In claude_worker.py around lines 171 to 180, the argument range check is done
manually; replace it by letting argparse validate the range by adding
choices=range(1, 11) to parser.add_argument("--id", ...) and remove the
subsequent if args.id < 1 or args.id > 10: ... block; update the help text if
desired to reflect the enforced range.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In config/config.example.yaml around lines 54 to 67, the config references

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856184

- [review_comment] 2025-09-16T23:20:21Z by coderabbitai[bot] (config/config.example.yaml:67)

```text
In config/config.example.yaml around lines 54 to 67, the config references
"{tenant}" in key_pattern/hash_key_pattern but never defines how "tenant" is
derived or configured; update the example and comments to explicitly define
"tenant" (e.g., per-application tenant ID, header-derived value, or environment
variable), show the exact configuration option name used to set it (or how to
derive it from request headers/metadata), and clarify its format/constraints;
also add a short note pointing to the relevant docs page (insert documentation
link placeholder) and add a docs cross-reference (README or operator guide) that
explains tenant resolution, recommended defaults, and examples for single-tenant
vs multi-tenant usage.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In config/config.example.yaml around lines 72 to 75, the example DSN includes a

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856191

- [review_comment] 2025-09-16T23:20:21Z by coderabbitai[bot] (config/config.example.yaml:75)

```text
In config/config.example.yaml around lines 72 to 75, the example DSN includes a
fake but parseable credential which can trigger scanners and encourage bad
habits; replace the DSN value with a non-parsable placeholder (e.g. an empty
string or clearly non-credential placeholder like "<DSN_HERE>") and remove any
fake username/password, and add a commented example environment variable entry
(OUTBOX_DSN) showing how to supply the DSN via env with a note marking it as
secret (e.g. "# OUTBOX_DSN (secret):
postgresql://user:password@host:port/db?sslmode=... — DO NOT COMMIT real
credentials"); ensure the file contains no parseable fake secrets.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In create_review_tasks.py around line 6 (and also update annotations at 53-54

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856199

- [review_comment] 2025-09-16T23:20:22Z by coderabbitai[bot] (create_review_tasks.py:6)

```text
In create_review_tasks.py around line 6 (and also update annotations at 53-54
and 70-75): replace the legacy typing.List import and usages with PEP 585 native
generics. Change the import to use Iterable from collections.abc (remove
typing.List), then update all type annotations to use built-in generics (e.g.,
list[str] instead of List[str], tuple[int, str] and Iterable[str] using the
modern syntax). Ensure any "typing.Tuple"/"typing.List" references are converted
to tuple[...] and list[...] and remove unused typing imports.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In create_review_tasks.py around lines 35 to 46, fromisoformat() currently fails

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856203

- [review_comment] 2025-09-16T23:20:22Z by coderabbitai[bot] (create_review_tasks.py:46)

```text
In create_review_tasks.py around lines 35 to 46, fromisoformat() currently fails
on ISO8601 strings that end with 'Z'; update parse_timestamp to normalize 'Z' to
an explicit offset (e.g. replace a trailing 'Z' with '+00:00') before calling
datetime.fromisoformat, keep the existing logic to set timezone to UTC when
missing, and wrap the fromisoformat call in a try/except that raises a clear
ValueError (e.g. "Invalid timestamp: <value>") if parsing still fails so the CLI
surfaces a helpful error message.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In create_review_tasks.py around lines 160 to 162, the file is opened without an

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856204

- [review_comment] 2025-09-16T23:20:22Z by coderabbitai[bot] (create_review_tasks.py:164)

```text
In create_review_tasks.py around lines 160 to 162, the file is opened without an
explicit encoding which can cause platform-dependent issues; update the open
call to specify UTF-8 and preserve Unicode by using: open(filename, "w",
encoding="utf-8"), and pass ensure_ascii=False to json.dump (e.g.,
json.dump(task, f, indent=2, ensure_ascii=False)) so non-ASCII characters are
written correctly.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In demos/lipgloss-transformation.tape around lines 22 to 43, the test currently

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856209

- [review_comment] 2025-09-16T23:20:22Z by coderabbitai[bot] (demos/lipgloss-transformation.tape:43)

```text
In demos/lipgloss-transformation.tape around lines 22 to 43, the test currently
sends many individual "Type"/"Enter" steps for static text which is slow and
brittle; replace that sequence with a single Paste/heredoc block: combine the
repeated Type/Enter lines into one heredoc payload (start a cat << 'EOF' block,
include the static lines in one paste, then close with EOF) so the tape sends
the whole static input in one step; remove the extra individual Type/Enter
entries and use the single Paste/heredoc step to improve speed and robustness.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In dependency_analysis.py around lines 326 to 338, validate_dependencies is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856214

- [review_comment] 2025-09-16T23:20:22Z by coderabbitai[bot] (dependency_analysis.py:339)

```text
In dependency_analysis.py around lines 326 to 338, validate_dependencies is
redundantly resolving aliases again even though get_normalized_feature_map()
already returns features with aliases resolved; remove the second resolve_alias
call inside the loop and use the dependency value from meta directly (or, if you
prefer explicit naming, rename the loop variable to normalized_dep) when
checking membership against feature_names and infrastructure_names so aliases
aren’t double-resolved.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/admin-api/monitoring.yaml around lines 1-66, this file duplicates

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856215

- [review_comment] 2025-09-16T23:20:22Z by coderabbitai[bot] (deployments/admin-api/monitoring.yaml:66)

```text
In deployments/admin-api/monitoring.yaml around lines 1-66, this file duplicates
canonical admin-api alerts and is missing required metadata.labels/ownership and
uses the wrong namespace; either remove this duplicate file or align it to the
canonical manifest by changing namespace to work-queue, adding metadata.labels
for the PrometheusRule (prometheus: kube-prometheus, role: alert-rules, app:
admin-api, component: admin-api) and for any ConfigMap (app: admin-api,
grafana_dashboard: "1"), and ensure every alert rule includes labels component:
admin-api so the Operator/Grafana can discover them.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Converted alert ConfigMap into a PrometheusRule resource. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Converted alert ConfigMap into a PrometheusRule resource. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_003.md:834
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/admin-api/monitoring.yaml around lines 11 to 18, the alert

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856220

- [review_comment] 2025-09-16T23:20:23Z by coderabbitai[bot] (deployments/admin-api/monitoring.yaml:18)

```text
In deployments/admin-api/monitoring.yaml around lines 11 to 18, the alert
expression only checks individual target up==0 and will not fire when all
targets vanish; replace the expr with a sum/absent check such as: use
sum(up{job="admin-api"}) == 0 or absent(up{job="admin-api"}) so the alert
triggers when the total is zero or the metric is missing, leaving the
for/labels/annotations unchanged.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/admin-api/monitoring.yaml around lines 29 to 36, the PromQL uses

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856225

- [review_comment] 2025-09-16T23:20:23Z by coderabbitai[bot] (deployments/admin-api/monitoring.yaml:36)

```text
In deployments/admin-api/monitoring.yaml around lines 29 to 36, the PromQL uses
histogram_quantile directly on per-series buckets which causes noisy per-series
quantiles; aggregate the bucket counts with sum by (le) (and any other desired
grouping like job/handler) over the rate window before passing to
histogram_quantile. Change the expression to sum the rate of
http_request_duration_seconds_bucket by (le) and then call
histogram_quantile(0.95, ...) so the 95th percentile is computed across
aggregated buckets rather than per-series.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/admin-api/monitoring.yaml around lines 73-76 (and also lines

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856229

- [review_comment] 2025-09-16T23:20:23Z by coderabbitai[bot] (deployments/admin-api/monitoring.yaml:84)

```text
In deployments/admin-api/monitoring.yaml around lines 73-76 (and also lines
125-127), the dashboard JSON is nested under a top-level "dashboard" key but
Grafana expects the dashboard object at the root; move the inner object out so
the root contains the dashboard fields directly, and add basic metadata keys for
import parity (schemaVersion, version, time, uid) at the root of that object;
ensure the final YAML places the dashboard object at the file root without the
extra "dashboard" wrapper and includes the recommended metadata fields.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/09_requirements.md around lines 47 to 49, the metrics list needs

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856235

- [review_comment] 2025-09-16T23:20:23Z by coderabbitai[bot] (docs/09_requirements.md:49)

```text
In docs/09_requirements.md around lines 47 to 49, the metrics list needs
explicit queue labeling and consistent Prometheus naming conventions: change the
queue_length entry to explicitly show the label (e.g.
queue_length{queue="<name>"}) to prevent accidental high cardinality, and ensure
all metric names use the correct suffixes and types (use _total for counters
like worker_registered_total and rate_limit_hits_total, _seconds for histograms
like job_processing_duration_seconds, and gauge for current values). Update the
text to explicitly state the required label and naming rules so automated tests
and implementers enforce queue label usage and consistent _total/_seconds
suffixes.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Acceptance criteria now spell out metric names, readiness semantics, and admin safeguards. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Acceptance criteria now spell out metric names, readiness semantics, and admin safeguards. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_002.md:931
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In docs/14_ops_runbook.md around lines 24 to 30, the local build example

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856245

- [review_comment] 2025-09-16T23:20:23Z by coderabbitai[bot] (docs/14_ops_runbook.md:30)

```text
In docs/14_ops_runbook.md around lines 24 to 30, the local build example
incorrectly uses --push (should use --load for local images) and the GO_VERSION
build-arg is inconsistent with the repository Dockerfiles (root Dockerfile line
3 uses FROM golang:1.23 and does not consume ARG GO_VERSION while go.mod and
other deployment Dockerfiles target Go 1.25); change the example flag from
--push to --load and reconcile the GO_VERSION mismatch by either adding ARG
GO_VERSION to the root Dockerfile and using it in the FROM (e.g., FROM
golang:${GO_VERSION}) so the build-arg takes effect, or update the
docs/build-arg value to the actual base image version used (1.23) so they match.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In internal/config/config.go (around lines 154–160) and docs/14_ops_runbook.md

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856247

- [review_comment] 2025-09-16T23:20:23Z by coderabbitai[bot] (docs/14_ops_runbook.md:48)

```text
In internal/config/config.go (around lines 154–160) and docs/14_ops_runbook.md
(lines 44–48): the docs claim double-underscore maps to nested keys but
config.go currently uses strings.NewReplacer(".", "_") with v.AutomaticEnv(),
which only maps dots to single underscores so CIRCUIT_BREAKER__COOLDOWN_PERIOD
will not resolve. Fix by either (A) code change: update the env key replacer to
translate double-underscores back to dots (and optionally also handle
single-underscore mapping) so ENV keys like CIRCUIT_BREAKER__COOLDOWN_PERIOD map
to circuit_breaker.cooldown_period, or (B) docs change: remove the
double-underscore example and explicitly document the actual mapping (e.g.,
WORKER_COUNT → worker.count, REDIS_ADDR → redis.addr) and parsing rules for
booleans/durations; apply one of these fixes and update tests/docs accordingly.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/dlq-remediation-pipeline.md around lines 168-176, the example

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856251

- [review_comment] 2025-09-16T23:20:24Z by coderabbitai[bot] (docs/api/dlq-remediation-pipeline.md:176)

```text
In docs/api/dlq-remediation-pipeline.md around lines 168-176, the example
response uses stringified duration values (e.g., "125ms") and is missing
idempotency guidance for write endpoints; replace duration strings with a
numeric duration_ms field (integer milliseconds) in all response examples
(including the other occurrences at 465-472 and 898-909) and update the POST
/pipeline/process-batch documentation to require/describe the Idempotency-Key
header on writes (explain header name, purpose, and that identical keys prevent
duplicate processing). Ensure examples and schema use duration_ms integers
consistently and add a short sentence in the process-batch endpoint docs
clarifying idempotency behavior.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/dlq-remediation-pipeline.md around lines 313-319 (and also apply the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856256

- [review_comment] 2025-09-16T23:20:24Z by coderabbitai[bot] (docs/api/dlq-remediation-pipeline.md:324)

```text
In docs/api/dlq-remediation-pipeline.md around lines 313-319 (and also apply the
same changes at 355-358 and 760-787), the matcher grammar sections mix the
structured JSON schema with free-form string examples; update the document to
consistently present only the structured matcher schema (error_pattern,
job_type, retry_count, optional time_window) and remove all ad-hoc/free-form
string examples in the "Update Rule" and "Patterns" sections, and add a short
note stating that free-form strings are deprecated and will be rejected by
validation so clients must use the structured schema.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/dlq-remediation-ui.md around lines 243 to 244, the example contains

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856258

- [review_comment] 2025-09-16T23:20:24Z by coderabbitai[bot] (docs/api/dlq-remediation-ui.md:244)

```text
In docs/api/dlq-remediation-ui.md around lines 243 to 244, the example contains
a raw JWT blob; replace the token value with a placeholder string (e.g.,
"<jwt>") so the JSON remains valid but no real token is shown, updating the
"confirmation_token" field to use the placeholder and keeping the rest of the
example unchanged.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/dlq-remediation-ui.md around lines 315-387 the API surface is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856260

- [review_comment] 2025-09-16T23:20:24Z by coderabbitai[bot] (docs/api/dlq-remediation-ui.md:387)

```text
In docs/api/dlq-remediation-ui.md around lines 315-387 the API surface is
described with ad‑hoc Markdown tables rather than a formal OpenAPI contract;
create and publish an openapi.yaml (e.g., docs/api/openapi.yaml) that defines
all schemas shown (DLQEntry, ErrorDetails, JobMetadata, AttemptRecord,
ErrorPattern, BulkOperationResult, OperationError) and endpoints, include
explicit enums for fields like ErrorPattern.severity (low, medium, high,
critical) and any role enums, and add request/response schemas; then wire
server-side validation middleware (e.g., ajv/express-openapi-validator or
framework equivalent) to enforce the contract for incoming requests and outgoing
responses, update docs to reference the openapi.yaml, and commit both the
openapi.yaml and the validation integration so the Markdown tables become a
generated/derived view from the authoritative OpenAPI file.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/SLAPS/FINAL-POSTMORTEM.md around line 20, the phrase "Zero

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856266

- [review_comment] 2025-09-16T23:20:24Z by coderabbitai[bot] (docs/SLAPS/FINAL-POSTMORTEM.md:20)

```text
In docs/SLAPS/FINAL-POSTMORTEM.md around line 20, the phrase "Zero
infrastructure" is misleading; replace it with "No distributed infra components"
(or equivalent) and update the surrounding sentence to reflect that local
filesystem operations plus tools like Git, directory watching, JSON tooling, and
Prometheus were used — not literally zero infrastructure. Search the document
for other occurrences of "Zero infrastructure" and update them consistently,
keeping the original meaning and tone.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/SLAPS/FINAL-POSTMORTEM.md around lines 50 to 51, the claim "WCAG

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856270

- [review_comment] 2025-09-16T23:20:24Z by coderabbitai[bot] (docs/SLAPS/FINAL-POSTMORTEM.md:51)

```text
In docs/SLAPS/FINAL-POSTMORTEM.md around lines 50 to 51, the claim "WCAG
accessibility compliance" is vague; update the sentence to specify the exact
WCAG version and conformance level (e.g., "WCAG 2.2 Level AA") and list which
surfaces were audited (e.g., Theme Playground UI, built-in themes, color
contrast, keyboard navigation, ARIA attributes, and API responses). Replace the
generic phrase with a parenthetical or short clause like "(WCAG 2.2 AA; audited:
Theme Playground UI, six built-in themes for color contrast and keyboard/ARIA,
and public theme API endpoints)" and, if applicable, add a brief note on the
method (automated + manual audit) and date of audit.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/SLAPS/FINAL-POSTMORTEM.md around lines 268 to 273, the resource metrics

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856271

- [review_comment] 2025-09-16T23:20:24Z by coderabbitai[bot] (docs/SLAPS/FINAL-POSTMORTEM.md:273)

```text
In docs/SLAPS/FINAL-POSTMORTEM.md around lines 268 to 273, the resource metrics
(15GB RAM, 78% CPU, load avg >20, multiple 4.5-hour rate-limiting pauses, 10
parallel developers) lack provenance; update those lines to either (A) attach
how and when each metric was measured (tool/command, metric source, exact
timestamps or time ranges) and include links or references to the raw
logs/monitoring screenshots/dashboards and any aggregation queries used, or (B)
remove or convert the numbers to qualitative statements if provenance cannot be
provided; ensure each retained metric has a clear source line (e.g., "measured
via Prometheus node exporter, 2025-09-10 02:00–06:30 UTC, see Grafana dashboard
link") so reviewers can verify.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/SLAPS/FINAL-POSTMORTEM.md around lines 421 to 472, the Appendix contains

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856273

- [review_comment] 2025-09-16T23:20:24Z by coderabbitai[bot] (docs/SLAPS/FINAL-POSTMORTEM.md:472)

```text
In docs/SLAPS/FINAL-POSTMORTEM.md around lines 421 to 472, the Appendix contains
high-level metrics presented without evidence; either remove or substantiate the
claims. Replace the standalone numbers with links to the concrete artifacts
(task lists, raw logs, LOC report, documentation index, API spec list, test
coverage report) or add footnotes/appendix subsections that embed/attach those
artifacts (paths, repo commits, CI artifacts, or generated reports), and if you
cannot provide verifiable sources remove the approximate figures and rewrite the
section to a qualitative summary without specific counts.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/SLAPS/FINAL-POSTMORTEM.md around lines 433-434, the statement "Test

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856278

- [review_comment] 2025-09-16T23:20:24Z by coderabbitai[bot] (docs/SLAPS/FINAL-POSTMORTEM.md:434)

```text
In docs/SLAPS/FINAL-POSTMORTEM.md around lines 433-434, the statement "Test
Coverage Achieved: 85%+ average" is ambiguous; replace it with a precise,
verifiable sentence that specifies the scope and measurement — list the
packages/modules included (placeholder: [PACKAGES]), the exact commit hash used
for the measurement (placeholder: [COMMIT_HASH]), the exact test/coverage
command run (placeholder: [COMMAND]), the coverage metric type
(line/statement/branch) and threshold (placeholder: [METRIC_TYPE] [THRESHOLD]),
the date/time of the run (placeholder: [DATE]), and a link or path to the
coverage artifact/report (placeholder: [ARTIFACT_URL_OR_PATH]); format it as a
single clear sentence using those placeholders so maintainers can replace them
with real values.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In README.md around lines 83-86 (and also lines 47-49 and 53-56), the examples

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856281

- [review_comment] 2025-09-16T23:20:25Z by coderabbitai[bot] (README.md:92)

```text
In README.md around lines 83-86 (and also lines 47-49 and 53-56), the examples
run/build Go commands without ensuring modules are fetched, which causes "module
not found" errors for new users; update the instructions to run "go mod
download" before any "go run" or "make build" examples (or explicitly note that
the Makefile performs module fetching) so users fetch dependencies first and
avoid first-run failures.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In README.md around lines 89 to 92, the example assumes bin/ exists so the go

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353856283

- [review_comment] 2025-09-16T23:20:25Z by coderabbitai[bot] (README.md:98)

```text
In README.md around lines 89 to 92, the example assumes bin/ exists so the go
build -o bin/tui command will fail if it doesn't; update the instructions to
create the directory first (e.g., run mkdir -p bin) or use a build step that
ensures the output directory exists before running go build, then run the binary
as shown.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/kubernetes/monitoring.yaml around lines 49-56, the alert expr

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2353924984

- [review_comment] 2025-09-17T00:25:55Z by coderabbitai[bot] (deployments/kubernetes/monitoring.yaml:56)

```text
In deployments/kubernetes/monitoring.yaml around lines 49-56, the alert expr
only matches when an individual target reports up==0 and will miss the case
where all targets disappear; change the expr to cover both absence and zero sum,
e.g. replace the current expr with a compound that uses sum and absent such as:
absent(up{app="admin-api"}) OR sum(up{app="admin-api"}) == 0, leaving the
for/labels/annotations intact so the alert fires when the app is fully missing
or all instances are down.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In BUGS.md around line 41, the note warns that the current "short block per

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912392

- [review_comment] 2025-09-18T12:12:35Z by coderabbitai[bot] (BUGS.md:74)

```text
In BUGS.md around line 41, the note warns that the current "short block per
queue in priority order" can starve low-priority work; replace that vague
instruction with a concrete weighted round‑robin algorithm using per‑priority
token buckets (example weights 8:2:1 for High:Med:Low), describe the refill
logic (reset bucket to weight when zero and decrement on each claim), require
weights configurable and validated as >0, and instruct adding per‑priority
metrics (tokens, claims, starve counters) so behavior is observable.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## BUGS.md around lines 62-64: the comment asks to "wire exactly-once" but lacks

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912408

- [review_comment] 2025-09-18T12:12:35Z by coderabbitai[bot] (BUGS.md:124)

```text
BUGS.md around lines 62-64: the comment asks to "wire exactly-once" but lacks
concrete handler integration points and a clear contract; update the doc to
explicitly spell out the handler contract and example steps: 1) derive an
idempotency key from the task, 2) call IdempotencyManager.Begin(ctx, key,
WithTTL) and short-circuit (return no-op) when it reports a duplicate or error,
3) ensure IdempotencyManager.End(ctx, key) is deferred/always-run after Begin
succeeds, 4) perform side-effects only after Begin succeeds and wrap any emitted
events in the outbox transaction/Enqueue so they are persisted atomically with
the handler, 5) on success mark the idempotency entry done, and 6) record
metrics idem_hits, idem_misses and outbox_flush_failures; replace the loose
suggestion with this precise step-by-step handler contract and note where to
increment metrics.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In BUGS.md around lines 65-66, the note currently lists BLMOVE as a

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912425

- [review_comment] 2025-09-18T12:12:35Z by coderabbitai[bot] (BUGS.md:126)

```text
In BUGS.md around lines 65-66, the note currently lists BLMOVE as a
"nice-to-have"; change the documentation and implementation guidance to make
BLMOVE the default for Redis ≥6.2 with a runtime fallback to BRPOPLPUSH when a
feature-probe or capability check fails. Update the text to instruct: perform a
Redis version or command-probing check at startup or before use; if BLMOVE is
available, use it by default; if the probe indicates absence or fails,
automatically fall back to BRPOPLPUSH and log a clear warning that the legacy
command is being used.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In cmd/job-queue-system/main.go around lines 4-13 and also apply to lines 53-58,

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912467

- [review_comment] 2025-09-18T12:12:35Z by coderabbitai[bot] (cmd/job-queue-system/main.go:13)

```text
In cmd/job-queue-system/main.go around lines 4-13 and also apply to lines 53-58,
the flag values are not normalized which lets values like "Admin" vs "admin"
cause bugs; add the "strings" import and after flag.Parse() trim spaces and
convert relevant flag variables to lowercase (e.g., flagVar =
strings.ToLower(strings.TrimSpace(flagVar))) for each flag that affects
behavior, ensuring all flag usages thereafter use the normalized variables.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In cmd/job-queue-system/main.go around lines 100 to 114, the signal handling

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912487

- [review_comment] 2025-09-18T12:12:36Z by coderabbitai[bot] (cmd/job-queue-system/main.go:114)

```text
In cmd/job-queue-system/main.go around lines 100 to 114, the signal handling
goroutine never calls signal.Stop, leaking signal notifications; after receiving
the first signal and calling cancel(), call signal.Stop(sigCh) to unregister the
channel from the signal package, then proceed with the existing logic to either
handle a second signal or wait the timeout and return so the goroutine exits and
resources are released.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Admin runs no longer spawn the background queue length updater. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Admin runs no longer spawn the background queue length updater. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:353
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In create_postmortem_tasks.py around lines 5, 18-19, 71-73, you are constructing

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912504

- [review_comment] 2025-09-18T12:12:36Z by coderabbitai[bot] (create_postmortem_tasks.py:5)

```text
In create_postmortem_tasks.py around lines 5, 18-19, 71-73, you are constructing
naive datetimes and appending "Z" manually; replace those with timezone-aware
UTC datetimes. Import timezone from datetime (or use datetime.timezone) and
replace datetime.now() (or naive constructions) with datetime.now(timezone.utc)
(or attach tzinfo=timezone.utc), then produce a proper Zulu-formatted string
either via .isoformat().replace('+00:00','Z') or format with
strftime('%Y-%m-%dT%H:%M:%SZ'); update all three locations accordingly so
timestamps are real UTC rather than naive times with a fake "Z".
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In create_postmortem_tasks.py around lines 134 to 142, the JSON files are opened

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912526

- [review_comment] 2025-09-18T12:12:36Z by coderabbitai[bot] (create_postmortem_tasks.py:142)

```text
In create_postmortem_tasks.py around lines 134 to 142, the JSON files are opened
without an explicit encoding and json.dump is left to default ASCII-escaping;
update the two open() calls to specify encoding='utf-8' and call json.dump(...,
ensure_ascii=False, indent=2) so the files are written deterministically in
UTF-8 and non-ASCII characters are not escaped.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/advanced-rate-limiting-api.md around lines 74–84, the example shows

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912542

- [review_comment] 2025-09-18T12:12:36Z by coderabbitai[bot] (docs/api/advanced-rate-limiting-api.md:84)

```text
In docs/api/advanced-rate-limiting-api.md around lines 74–84, the example shows
sleeping once then returning an error; change it to demonstrate a capped retry
loop with backoff and proper cancellation: introduce a package-level
ErrRateLimited for callers to check, then replace the single sleep/return with a
loop that attempts rl.Consume up to a maxRetries, uses result.RetryAfter (or an
exponential backoff capped to a maxDelay) between attempts, respects ctx
cancellation/deadline, and returns ErrRateLimited if retries are exhausted (or
the context is done) so callers can handle rate-limit errors explicitly.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/advanced-rate-limiting-api.md around lines 140 to 158, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912558

- [review_comment] 2025-09-18T12:12:36Z by coderabbitai[bot] (docs/api/advanced-rate-limiting-api.md:158)

```text
In docs/api/advanced-rate-limiting-api.md around lines 140 to 158, the
MinGuaranteedShare description and guardrail are contradictory; adopt Model B
(weight‑proportional minima): rename or clarify the field as a total
MinGuaranteedBudget (0.0–1.0) that is distributed per priority as minimum_i =
(weight_i / Σweights) * MinGuaranteedBudget, update the guardrail to require
MinGuaranteedBudget ≤ 1.0, document the per‑priority calculation, and keep the
existing renormalisation behavior and warning log semantics (cap values at 1.0
and clamp negatives to 0) so operators know what will happen if the budget
exceeds capacity.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/advanced-rate-limiting-api.md around lines 214–221 the field name

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912577

- [review_comment] 2025-09-18T12:12:36Z by coderabbitai[bot] (docs/api/advanced-rate-limiting-api.md:221)

```text
In docs/api/advanced-rate-limiting-api.md around lines 214–221 the field name
DryRunWouldAllow is ambiguous; update the documentation to explicitly state that
this boolean means "true if the request would have been allowed when
DryRun=false (i.e., in non-dry-run mode)". Alternatively, if you prefer a
clearer identifier, rename the field in the code and docs to AllowedIfNotDryRun
(or WouldHaveBeenAllowedIfNotDryRun), update all references and API consumers,
and ensure the doc comment matches the new name and semantics.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/advanced-rate-limiting-api.md around lines 223 to 236, the Status

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912602

- [review_comment] 2025-09-18T12:12:37Z by coderabbitai[bot] (docs/api/advanced-rate-limiting-api.md:236)

```text
In docs/api/advanced-rate-limiting-api.md around lines 223 to 236, the Status
struct mixes scope-level and priority-level fields (Priority/Weight) creating
ambiguity about the contract for GetStatus(ctx, scope). Either remove Priority
and Weight from Status and introduce a separate FairnessStatus (or
PriorityStatus) type and update examples to call GetFairnessStatus/GetStatus as
appropriate, or document that Status represents a (scope, priority) tuple by
renaming the type to StatusForPriority and updating method signatures/examples
to accept/return a priority-scoped status; pick one approach and make the
corresponding API doc changes consistently (type name, field list, example
calls, and method description).
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/advanced-rate-limiting-api.md around lines 365 to 375, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912613

- [review_comment] 2025-09-18T12:12:37Z by coderabbitai[bot] (docs/api/advanced-rate-limiting-api.md:375)

```text
In docs/api/advanced-rate-limiting-api.md around lines 365 to 375, the
"Configure TTLs" best-practice is too vague; replace or augment bullet 5 with a
concrete TTL rule-of-thumb: add a new bullet 5 stating "KeyTTL >=
max(2×RefillInterval, 2×BurstSize/RatePerSecond) and never set below 2× the
longest expected idle gap (to avoid bucket evaporation and cold-start spikes)".
Keep numbering of subsequent items, ensure formatting matches existing bullets,
and keep the language concise.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/canary-deployments.md around lines 20 to 23, the docs state the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912669

- [review_comment] 2025-09-18T12:12:37Z by coderabbitai[bot] (docs/api/canary-deployments.md:25)

```text
In docs/api/canary-deployments.md around lines 20 to 23, the docs state the
content type but do not declare the global timestamp format; add a single clear
sentence after the content-type line that declares all timestamps follow RFC
3339 in UTC (Z) so clients know the canonical format, and update any example
request/response timestamps in this document to use that RFC 3339 UTC format for
consistency.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/canary-deployments.md around lines 40 to 49, the error codes table

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912707

- [review_comment] 2025-09-18T12:12:37Z by coderabbitai[bot] (docs/api/canary-deployments.md:51)

```text
In docs/api/canary-deployments.md around lines 40 to 49, the error codes table
is missing authentication/authorization entries; add two rows to the table: one
for the 401 case (e.g., code `UNAUTHENTICATED` or `UNAUTHORIZED` with
description like "Authentication required" and HTTP Status `401`) and one for
the 403 case (e.g., code `FORBIDDEN` with description like "Insufficient
permissions" and HTTP Status `403`), ensuring they follow the same table
formatting as the existing rows.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## docs/api/canary-deployments.md lines 68-69 (and also update occurrences at

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912737

- [review_comment] 2025-09-18T12:12:38Z by coderabbitai[bot] (docs/api/canary-deployments.md:71)

```text
docs/api/canary-deployments.md lines 68-69 (and also update occurrences at
366-371, 373-378, 779-785, 560-566, 594-595): the spec currently declares ULIDs
(uppercase, 26 chars) but examples use a prefixed UUID ("canary_..."), causing a
contradiction; pick one format and make amendments: either (A) adopt plain
uppercase 26-char ULIDs everywhere — remove the "canary_" prefix from all
example IDs and ensure any descriptive text and regex examples reflect 26
uppercase ULID characters, or (B) keep the "canary_" prefix — update the spec
text to state "prefix + ULID" and adjust any regex/validation examples to accept
the literal prefix followed by a 26-char uppercase ULID; apply the chosen change
consistently to the listed line ranges and any related ID examples in the
document.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/canary-deployments.md around lines 100 to 116 (and also apply same

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912759

- [review_comment] 2025-09-18T12:12:38Z by coderabbitai[bot] (docs/api/canary-deployments.md:118)

```text
In docs/api/canary-deployments.md around lines 100 to 116 (and also apply same
changes to 118-130, 507-553, 573-588), the request example uses inconsistent
field names and duration formats that drift from the canonical Deployment.config
schema; rename max_duration/min_duration to
max_canary_duration/min_canary_duration (or vice-versa to match
Deployment.config exactly), normalize all duration values to the canonical
format (e.g., "5m0s" rather than "5m"), and update the Parameters sections
mentioned to use the same field names and canonical duration format so clients
see a single consistent public schema across the document.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/canary-deployments.md around lines 110-113 (also apply fixes at

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912771

- [review_comment] 2025-09-18T12:12:38Z by coderabbitai[bot] (docs/api/canary-deployments.md:115)

```text
In docs/api/canary-deployments.md around lines 110-113 (also apply fixes at
512-517, 529-537, 543-552), the duration values use mixed formats like "2h" and
"5m" rather than canonical Go time.Duration strings; normalize all duration
fields (e.g., max_duration, min_duration, metrics_window, and any other duration
keys) to full Go canonical form such as "2h0m0s" and "5m0s" consistently across
the file, updating the example JSON/YAML values and any explanatory text so
every duration uses the same canonical format.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/canary-deployments.md around lines 146 to 162, the request body for

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912780

- [review_comment] 2025-09-18T12:12:38Z by coderabbitai[bot] (docs/api/canary-deployments.md:164)

```text
In docs/api/canary-deployments.md around lines 146 to 162, the request body for
the PUT /deployments/{id}/percentage endpoint does not define the numeric type,
valid bounds or validation/rounding behavior; update the docs to state that
"percentage" is a required number between 0 and 100 (inclusive), specify whether
integers and decimals are accepted (e.g., allow decimals to one or two decimal
places and preserve float precision), define handling for edge cases (reject
NaN, +Inf, -Inf; cap or reject >100 or <0 according to API policy — prefer
rejecting out-of-range values), and document validation/rounding rules (e.g.,
server validates and returns 400 for invalid values, or rounds to two decimal
places if automatic rounding is applied). Also add the error response shape and
status code for validation failures (e.g., 400 with JSON body containing
error_code, message, and field errors array detailing the "percentage" issue) so
callers know expected validation behavior.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/canary-deployments.md around lines 263 to 268, the throughput

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912789

- [review_comment] 2025-09-18T12:12:38Z by coderabbitai[bot] (docs/api/canary-deployments.md:270)

```text
In docs/api/canary-deployments.md around lines 263 to 268, the throughput
message mixes English and math by saying "Throughput decrease: -5.1%"; update
the message generation so the sign and wording match: either format as
"Throughput decrease: 5.1%" (remove the negative sign when using the word
"decrease") or change the label to a neutral term like "Throughput change:
-5.1%" (keep the negative sign). Modify the template or formatting logic
accordingly so negative values drop the minus when using "decrease" or retain
the minus when using "change".
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/canary-deployments.md around lines 761–786, update the WebSocket

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912812

- [review_comment] 2025-09-18T12:12:38Z by coderabbitai[bot] (docs/api/canary-deployments.md:786)

```text
In docs/api/canary-deployments.md around lines 761–786, update the WebSocket
section to use wss by default and add an Authentication subsection: replace the
ws:// URL with wss://, state that clients should send Authorization: Bearer
<token> during the WebSocket handshake as the preferred method, document support
for an optional ?token=<...> query param only if the server enables it, and
include a brief wscat example showing how to connect with an Authorization
header (e.g., wscat -c wss://... -H "Authorization: Bearer <token>") so readers
know how to pass the bearer token.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/capacity-planning-api.md around lines 16 to 29, the CapacityPlanner

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912838

- [review_comment] 2025-09-18T12:12:38Z by coderabbitai[bot] (docs/api/capacity-planning-api.md:29)

```text
In docs/api/capacity-planning-api.md around lines 16 to 29, the CapacityPlanner
interface references types (PlanRequest, PlanResponse, PlannerConfig,
PlannerState, context.Context) that are not defined or linked, making the API
doc incomplete; add definitions for each referenced type directly in this
document (struct fields, field types and optional descriptions) or add clear
links to the generated API/type reference for each type, and ensure any imported
types (e.g., context.Context) are noted with their package and a link; keep the
format consistent with the rest of the doc and update the table of contents or
header to reflect the added type definitions.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/capacity-planning-api.md around lines 289 to 308, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912871

- [review_comment] 2025-09-18T12:12:39Z by coderabbitai[bot] (docs/api/capacity-planning-api.md:308)

```text
In docs/api/capacity-planning-api.md around lines 289 to 308, the
TrafficSpike/TrafficPattern enum values are inconsistently named (docs say
“instant, linear, exp, bell” while examples use `SpikeBell`) and the actual
constant names are missing; update the doc to list the exact constant names used
in the code (include the PatternType and SpikeShape constant values as defined
earlier), replace informal names with the precise enum identifiers used in the
type defs, and add a short code block showing the constants so the documentation
and examples match exactly.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/capacity-planning-api.md around lines 365 to 373 the example calls

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912889

- [review_comment] 2025-09-18T12:12:39Z by coderabbitai[bot] (docs/api/capacity-planning-api.md:373)

```text
In docs/api/capacity-planning-api.md around lines 365 to 373 the example calls
planner.GeneratePlan and panics on error — we should not teach panicking in
examples; instead handle the error properly. Replace the panic with structured
error handling: if this is a main/demo show a graceful exit using log.Fatalf
with a clear message and the error (or os.Exit after logging), otherwise return
the error up the call stack (or wrap it with context and return). Ensure the
example imports/uses the chosen logger or returns the error so the sample
demonstrates safe, production-appropriate error handling.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/capacity-planning-api.md around lines 479 to 494, the code

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912898

- [review_comment] 2025-09-18T12:12:39Z by coderabbitai[bot] (docs/api/capacity-planning-api.md:494)

```text
In docs/api/capacity-planning-api.md around lines 479 to 494, the code
auto-applies scaling whenever Plan.Confidence >= config.ConfidenceThreshold
without checking whether the proposed plan will keep SLOs met; add an SLO gate
before auto-apply by computing/consulting a predicted SLO compliance check
(e.g., predictedSLOCompliant(response.Plan) or using
response.Plan.PredictedSLOCompliance) and only call applyScalingPlan when both
confidence >= threshold AND predicted SLO compliance is true; also log a clear
message when auto-apply is skipped due to SLO risk and surface the predicted SLO
metrics in the log for operators to inspect.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/capacity-planning-api.md around lines 508 to 561, the error-handling

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912912

- [review_comment] 2025-09-18T12:12:39Z by coderabbitai[bot] (docs/api/capacity-planning-api.md:561)

```text
In docs/api/capacity-planning-api.md around lines 508 to 561, the error-handling
example mixes direct type assertions with pointer/value forms and uses a bare
time.Sleep; replace the direct type assertion with errors.As to reliably extract
a *capacityplanning.PlannerError, and remove the magic time.Sleep by showing a
retry loop that respects context deadlines (e.g., loop with backoff and check
ctx.Done or a context.WithDeadline/WithTimeout) so the example demonstrates
safe, cancellable retries in library code.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/capacity-planning-api.md around lines 508 to 533, add a canonical

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912923

- [review_comment] 2025-09-18T12:12:39Z by coderabbitai[bot] (docs/api/capacity-planning-api.md:533)

```text
In docs/api/capacity-planning-api.md around lines 508 to 533, add a canonical
mapping table that translates each PlannerError.Code to the appropriate HTTP
status code and gRPC status code; list each error code (INVALID_METRICS,
INSUFFICIENT_HISTORY, FORECAST_FAILED, MODEL_NOT_SUPPORTED, CONFIG_INVALID,
SLO_UNACHIEVABLE, CAPACITY_LIMIT_EXCEEDED, COOLDOWN_ACTIVE, ANOMALY_DETECTED)
with a recommended HTTP status (e.g., 400 for client errors, 404/409 where
appropriate, 429 for rate/cooldown, 500 for server/forecast failures) and
corresponding gRPC canonical codes (e.g., INVALID_ARGUMENT, NOT_FOUND,
FAILED_PRECONDITION/ALREADY_EXISTS as applicable, RESOURCE_EXHAUSTED for limits,
UNAVAILABLE/INTERNAL for engine failures), and include a one-line rationale
column for each mapping to justify the choice.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/capacity-planning-api.md around lines 631-640, the complexity table

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912932

- [review_comment] 2025-09-18T12:12:39Z by coderabbitai[bot] (docs/api/capacity-planning-api.md:640)

```text
In docs/api/capacity-planning-api.md around lines 631-640, the complexity table
is too vague: update the table and surrounding text to list practical caps and
numerical-stability guards (e.g., for M/M/c state that complexity is O(min(c,
C_MAX)) and document a configurable cap C_MAX and checks to avoid numerical
instability when c is large), clarify Holt-Winters complexity as O(n * k * it)
or O(n * s) by specifying per-iteration constants (k = number of seasonal
components or smoothing parameters and it = number of iterations) and any
early‑stop/regularization applied, and add brief notes for Simulation and
Pattern Extraction about applied caps or downsampling (safeguards like max
steps, max history, or sampling) so the table reflects real-world limits rather
than idealized O()s.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/SLAPS/worker-reflections/claude-001-reflection.md lines 1-4, the file

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912941

- [review_comment] 2025-09-18T12:12:39Z by coderabbitai[bot] (docs/SLAPS/worker-reflections/claude-001-reflection.md:4)

```text
In docs/SLAPS/worker-reflections/claude-001-reflection.md lines 1-4, the file
lacks the YAML front-matter used by other reflections; add a YAML front-matter
block at the very top including a date (YYYY-MM-DD) and worker_id: claude-001
following the same field names and formatting as the other reflection files so
the document is consistent and parsable by the site generator.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/SLAPS/worker-reflections/claude-005-reflection.md around lines 1 to 3,

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358912986

- [review_comment] 2025-09-18T12:12:40Z by coderabbitai[bot] (docs/SLAPS/worker-reflections/claude-005-reflection.md:3)

```text
In docs/SLAPS/worker-reflections/claude-005-reflection.md around lines 1 to 3,
this markdown file is missing the YAML front-matter used by other reflections;
add a YAML block at the very top containing at least date and worker_id (e.g.,
date: YYYY-MM-DD and worker_id: claude-005) to match the pattern used by other
reflection files and ensure consistency across the docs.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 24 to 27, the integration test

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358913026

- [review_comment] 2025-09-18T12:12:40Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:27)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 24 to 27, the integration test
"TestWebhookDeliveryWithRetries" is incorrectly listed under unit tests for
signatures; remove that entry from the unit-test list and add it to the Webhook
Harness (integration) section or note it as a separate integration scenario.
Update the documentation to either split the test into a unit and integration
entry or relocate the integration entry, and adjust headings/bullet grouping so
unit tests only contain true unit tests and integration scenarios appear under
the Webhook Harness section.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 218 to 225, the test commands

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358913042

- [review_comment] 2025-09-18T12:12:40Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:225)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 218 to 225, the test commands
do not enable the Go race detector (and optionally test shuffling) which misses
concurrency bugs; update the documented test commands to include the -race flag
on go test (e.g., go test -v -race ./... and go test -v -race
-coverprofile=coverage.out ./...) and optionally show adding test shuffling
(e.g., -shuffle=on) where supported, and update any examples or notes to mention
using -race (and -shuffle) for concurrency-sensitive suites.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 232 to 235, add an additional

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358913071

- [review_comment] 2025-09-18T12:12:41Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:235)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 232 to 235, add an additional
example command to show how to run only the payload redactor tests; duplicate
the existing security-only example and append a new line with the go test
command using the '^TestPayloadRedactor_' run selector so contributors can
easily execute only redactor tests.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 10 | Integration & security examples now rely on explicit package paths. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Integration & security examples now rely on explicit package paths. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_007.md:1030
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In AGENTS.md around line 3, remove the trailing whitespace at the end of the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358974163

- [review_comment] 2025-09-18T12:25:36Z by coderabbitai[bot] (AGENTS.md:3)

```text
In AGENTS.md around line 3, remove the trailing whitespace at the end of the
line "Quick notes for working on this repo (Go Redis Work Queue)" so the line
ends immediately after the closing parenthesis; update the file to save without
the extra space, verify no other trailing spaces exist on that line, and commit
the change.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In AGENTS.md around line 10, the heading "Table of Contents" is currently an H1

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358974184

- [review_comment] 2025-09-18T12:25:36Z by coderabbitai[bot] (AGENTS.md:10)

```text
In AGENTS.md around line 10, the heading "Table of Contents" is currently an H1
which duplicates the document's main H1; change it to an H2 (replace leading "#
" with "## "), ensure there is a blank line before and after the new H2 per
Markdown conventions, and verify the file retains only a single H1 (the document
title) elsewhere.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In AGENTS.md around lines 11 to 39 the TOC list blocks do not have blank lines

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358974196

- [review_comment] 2025-09-18T12:25:36Z by coderabbitai[bot] (AGENTS.md:39)

```text
In AGENTS.md around lines 11 to 39 the TOC list blocks do not have blank lines
before and after them, triggering markdownlint MD032/MD022; fix by inserting a
blank line immediately before the start of the TOC list and a blank line
immediately after the closing list block (and any nested list blocks),
preserving existing indentation and content so each list is separated by a
single empty line.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In AGENTS.md around lines 34 to 36, the TOC contains a link text that includes

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358974210

- [review_comment] 2025-09-18T12:25:36Z by coderabbitai[bot] (AGENTS.md:36)

```text
In AGENTS.md around lines 34 to 36, the TOC contains a link text that includes
literal heading hashes ("##### ") and a duplicated entry; edit the TOC so the
link text does not contain any heading hashes (remove the "##### " prefix) and
delete the duplicate line, leaving a single correctly formatted anchor link
(keep the anchor targets unchanged).
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In AGENTS.md around lines 566 to 594, headings inside the blockquote/admonition

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358974220

- [review_comment] 2025-09-18T12:25:37Z by coderabbitai[bot] (AGENTS.md:594)

```text
In AGENTS.md around lines 566 to 594, headings inside the blockquote/admonition
make fragile anchors because GitHub renders anchors inconsistently; move the
heading(s) out of the blockquote or add plain headings immediately after the
admonition so stable anchors exist (e.g., keep the admonition content but
duplicate the heading as a non-blockquote line right after, or convert the
blockquoted heading into a normal heading and keep the admonition content below)
to ensure TOC links and anchor targets resolve reliably.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/admin-api.md around lines 10-11 (also apply same change at 41-46 and

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358974241

- [review_comment] 2025-09-18T12:25:37Z by coderabbitai[bot] (docs/api/admin-api.md:11)

```text
In docs/api/admin-api.md around lines 10-11 (also apply same change at 41-46 and
179-186), the phrase "Double Confirmation" is misleading because it implies two
inputs when the implementation uses distinct confirmation phrases per endpoint;
update the wording to explicitly state that each dangerous endpoint requires a
unique confirmation phrase (or, if intended, change the API docs to show two
separate confirmation fields), and revise the examples to either list the
distinct phrase required per endpoint or show two confirmation fields so
operators are not confused during incidents.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/admin-api.md around lines 82-84 (and also at 99-106 and 139-146),

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358974260

- [review_comment] 2025-09-18T12:25:37Z by coderabbitai[bot] (docs/api/admin-api.md:84)

```text
In docs/api/admin-api.md around lines 82-84 (and also at 99-106 and 139-146),
the queue name examples in the TOC (e.g., "jobqueue:high") do not match the
mapping table (which uses "jobqueue:high_priority"); pick one canonical name and
make all examples consistent: update the TOC examples to use the mapping table
names OR rename the mapping entries to match the TOC, and then replace every
mismatched occurrence at the three specified ranges so the sample keys and the
mapping table are identical.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/admin-api.md around lines 149 to 177 (and likewise apply to lines

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358974269

- [review_comment] 2025-09-18T12:25:37Z by coderabbitai[bot] (docs/api/admin-api.md:177)

```text
In docs/api/admin-api.md around lines 149 to 177 (and likewise apply to lines
179 to 200), the minimum reason length (10 characters) is repeated inline; add a
single clear line such as "Minimum reason length: 10 characters" immediately
before each relevant endpoint description and update the purge-all reference to
point to that line so both endpoints reference the same minimum-length statement
rather than duplicating the rule.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/admin-api.md around lines 331 to 354, the Go example references a

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2358974289

- [review_comment] 2025-09-18T12:25:37Z by coderabbitai[bot] (docs/api/admin-api.md:354)

```text
In docs/api/admin-api.md around lines 331 to 354, the Go example references a
non‑existent StatsResponse type which will break copy‑paste; either add a
concise StatsResponse struct definition showing the expected fields (e.g.,
counts/timestamps/metrics returned by /api/v1/stats) and mention it above the
sample, or change the example to decode into a raw map[string]interface{} and
update the function signature/return type accordingly; also ensure the sample
handles the http.NewRequest error and returns the json.Decoder.Decode error
instead of ignoring it.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## `

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360032891

- [review_comment] 2025-09-18T15:55:16Z by coderabbitai[bot] (.claude/agents/feature-enhancer.md:14)

```text
`
In .claude/agents/feature-enhancer.md around lines 6 to 14 (and also apply same
fixes to ranges 22-33, 34-44, 80-136, 200-224, 252-282), fix markdown lint
violations: ensure there is a blank line before and after each heading, add a
blank line between list items and surrounding blocks, add an empty line before
and after fenced code blocks (
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## .claude/agents/feature-enhancer.md around lines 38 to 42: the file uses a

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360032904

- [review_comment] 2025-09-18T15:55:16Z by coderabbitai[bot] (.claude/agents/feature-enhancer.md:42)

```text
.claude/agents/feature-enhancer.md around lines 38 to 42: the file uses a
non-standard admonition "[!note]-" that won't render in portable Markdown;
replace it with a simple heading plus a blockquote. Change the admonition to a
heading like "### 🗣️ CLAUDE'S THOUGHTS" and move the instructional text into a
Markdown blockquote (prefix lines with "> "), ensuring the fenced code block
remains marked as markdown; update punctuation or dash characters as needed so
the output is valid Markdown and renders in common viewers.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In .claude/agents/feature-enhancer.md around lines 165 to 190, replace the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360032914

- [review_comment] 2025-09-18T15:55:16Z by coderabbitai[bot] (.claude/agents/feature-enhancer.md:190)

```text
In .claude/agents/feature-enhancer.md around lines 165 to 190, replace the
inline XML-commented hard-coded color block with a reusable reference to
docs/colors.yml (key: feature_enhancement_palette) and present the palette as a
concise Markdown table of Category / Token / Hex (e.g., Background: canvas ->
#0d1117, panel -> #161b22; Border: default -> #30363d, success -> #238636, error
-> #f85149, warning -> #ffa657; Text: primary -> #f0f6fc, secondary -> #8b949e,
link -> #58a6ff, success -> #39d353, error -> #f85149; Buttons: primary ->
#238636, secondary -> #30363d, info -> #58a6ff). Remove the XML comment block
and add a short sentence pointing readers to docs/colors.yml for the canonical
source and the Markdown table for quick reference.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In .githooks/pre-commit around lines 19 to 25, the script unconditionally runs

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360032927

- [review_comment] 2025-09-18T15:55:17Z by coderabbitai[bot] (.githooks/pre-commit:25)

```text
In .githooks/pre-commit around lines 19 to 25, the script unconditionally runs
"git add docs/features-ledger.md README.md" which force-stages changes without
an opt-out; modify the script to gate the auto-stage behind an environment flag
(for example AUTO_STAGE_DOCS) so that by default it auto-stages but if
AUTO_STAGE_DOCS is set to false/0 it skips staging; implement a simple
conditional around the git add that echoes a message when skipping and preserves
the existing "|| true" behavior to avoid breaking the hook.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In .gitignore around line 16, you added / .gocache/ but missed other common

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360032954

- [review_comment] 2025-09-18T15:55:17Z by coderabbitai[bot] (.gitignore:16)

```text
In .gitignore around line 16, you added / .gocache/ but missed other common
local/build artifacts; update the file to also ignore typical Go and editor/temp
files such as /bin/, /pkg/, /vendor/, *.exe, *.test, coverage.out, *.env and
.env.local, .DS_Store, and editor dirs like .idea/ and .vscode/ (or expand the
existing allowlist pattern), and ensure you do a quick sanity check that go.work
is not required by the repo before adding it to the ignore list.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Added `go.work.sum` to the ignore list alongside `go.work`. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Added `go.work.sum` to the ignore list alongside `go.work`. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:422
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## AGENTS.md lines 828-830: the second "APPENDIX B: WILD IDEAS — HAVE A BRAINSTORM"

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360032978

- [review_comment] 2025-09-18T15:55:17Z by coderabbitai[bot] (AGENTS.md:830)

```text
AGENTS.md lines 828-830: the second "APPENDIX B: WILD IDEAS — HAVE A BRAINSTORM"
heading is a duplicate and should be removed; delete this repeated heading (and
any immediately duplicated anchor/ID) so only the original Appendix B remains,
or if those lines introduce unique content, merge that content under the first
Appendix B heading and remove the duplicate anchor to ensure no duplicate
anchors exist.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In BUGS.md around lines 3-4 (also apply the same change to lines 51-53 and

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360032989

- [review_comment] 2025-09-18T15:55:17Z by coderabbitai[bot] (BUGS.md:4)

```text
In BUGS.md around lines 3-4 (also apply the same change to lines 51-53 and
71-72), the intro contradicts itself by claiming the heartbeat/reaper will SCAN
jobqueue:worker:*:processing keys while elsewhere advising readers to avoid
using KEYS/SCAN; update the text to be consistent with the registry +
single-slot guidance by removing any endorsement of SCAN/KEYS and instead
describing that the heartbeat/reaper uses the registry and single-slot ownership
to locate processing keys (or explain that any reaper must be implemented via
the registry and safe single-slot semantics), and rephrase the listed items to
reflect that SCAN/KEYS are discouraged and not part of the recommended
implementation.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In BUGS.md around lines 12–47: the heartbeat example blocks the worker (no

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033007

- [review_comment] 2025-09-18T15:55:17Z by coderabbitai[bot] (BUGS.md:47)

```text
In BUGS.md around lines 12–47: the heartbeat example blocks the worker (no
goroutine), renews without verifying ownership, and relies on a non‑standard
jitter package; replace it with: perform an initial SET NX with TTL (PX) to
claim the heartbeat, start a background goroutine that does jittered time.Sleep
using the standard library (math/rand + time) and on each tick runs an atomic
Lua script which checks the key value == workerID and only then extends the TTL
(so only owner can renew), handle transient errors with retry logic and
non‑blocking warning logs, cancel the goroutine via context cancel before final
LREM/DEL so it exits cleanly, and document the Lua script and standard library
jitter approach for portability.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In BUGS.md around lines 55 to 69 (and also line 71) the worker registry

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033018

- [review_comment] 2025-09-18T15:55:17Z by coderabbitai[bot] (BUGS.md:69)

```text
In BUGS.md around lines 55 to 69 (and also line 71) the worker registry
currently uses a plain SADD which never expires and lets dead worker IDs
accumulate; change the design so membership is self‑healing by either (a)
switching to a ZSET storing lastSeen timestamps and updating the member score on
each heartbeat so the reaper can ZREMRANGEBYSCORE (or ZRANGEBYSCORE to find
stale IDs) and remove entries older than a timeout, or (b) creating a per‑worker
key with a short TTL that the worker refreshes on heartbeat and having the
reaper only consider workers with an existing key (and remove any orphaned SADD
entries if you keep the set). Update reaper logic to use ZRANGEBYSCORE or check
TTLs instead of scanning the whole set so zombies are pruned automatically.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In BUGS.md around lines 119–121, the ledger guidance needs concrete failure-mode

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033024

- [review_comment] 2025-09-18T15:55:17Z by coderabbitai[bot] (BUGS.md:121)

```text
In BUGS.md around lines 119–121, the ledger guidance needs concrete failure-mode
and redaction requirements: update the doc to require emitting ack/history
events to a durable sink (e.g., S3/Kafka) while retaining the existing LREM
procList 1 payload after success; mandate a bounded non‑blocking local channel
(with drop counter/metrics) when the sink is unreachable and explicit
alerts/backpressure so workers fail fast instead of silently dropping history;
specify local fallback must use an atomic appender with daily rotation, gzip
compression, size caps, documented retention policy and retention enforcement,
and list default redactions (JWTs/tokens, emails, PII IDs) plus rotation caps
and scrub rules to guarantee redaction; also call out instrumentation/alerts and
backoff semantics to surface sink outages.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In cmd/admin-api/main.go lines 80-89, the missing-config check incorrectly uses

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033042

- [review_comment] 2025-09-18T15:55:18Z by coderabbitai[bot] (cmd/admin-api/main.go:89)

```text
In cmd/admin-api/main.go lines 80-89, the missing-config check incorrectly uses
os.IsNotExist on the error from v.ReadInConfig; Viper returns
viper.ConfigFileNotFoundError instead. Replace the os.IsNotExist check with a
type check for viper.ConfigFileNotFoundError (e.g., via errors.As) and only
return the default cfg, nil in that case; for all other errors from
ReadInConfig, return the error as-is.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In cmd/tui/main.go around lines 38 to 55, don't discard flag usage output and

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033071

- [review_comment] 2025-09-18T15:55:18Z by coderabbitai[bot] (cmd/tui/main.go:55)

```text
In cmd/tui/main.go around lines 38 to 55, don't discard flag usage output and
add a --version flag: remove or stop calling fs.SetOutput(io.Discard) so
help/usage is printed to the user (use the default or os.Stdout), add a new flag
(e.g., boolVar(&showVersion, "version", false, "Show version and exit")), and
after parsing check if showVersion is set and print the program version string
to stdout then exit; keep parse error handling but ensure normal --help and
--version both produce visible output.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In cmd/tui/main.go around lines 74 to 94, the current logic discards the host

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033095

- [review_comment] 2025-09-18T15:55:18Z by coderabbitai[bot] (cmd/tui/main.go:94)

```text
In cmd/tui/main.go around lines 74 to 94, the current logic discards the host
portion of metrics-addr and only stores the port, which breaks host-bound
addresses like 127.0.0.1:9090; instead preserve the full host:port when
metricsAddr contains a host, and only fall back to parsing a bare port if
metricsAddr has no colon. Set cfg.Observability.MetricsAddress (or the existing
config field) to the full metricsAddr when given, and only parse and set
MetricsPort if you need the numeric port separately; also update the
observability server startup to prefer using the full host:port value from the
config if present.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In cmd/tui/main.go around lines 109 to 112, the Redis Ping is using the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033107

- [review_comment] 2025-09-18T15:55:18Z by coderabbitai[bot] (cmd/tui/main.go:112)

```text
In cmd/tui/main.go around lines 109 to 112, the Redis Ping is using the
background context and can hang on dead networks; wrap the ping call in a short
cancellable context (e.g., context.WithTimeout(ctx, 2*time.Second)), defer
cancel(), then call rdb.Ping(timeoutCtx).Result() and handle the error as before
(including exiting on failure); ensure you import time if not already imported.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In dependency_analysis.py around lines 283 to 287 (and also the earlier mapping

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033126

- [review_comment] 2025-09-18T15:55:18Z by coderabbitai[bot] (dependency_analysis.py:287)

```text
In dependency_analysis.py around lines 283 to 287 (and also the earlier mapping
at ~line 147), there's a duplicate "storage-backends" key causing collisions;
keep the feature entry at line 285 and rename the infra mapping at line 147 to a
distinct name (e.g. storage_backends_runtime), then update all references/usages
throughout the codebase to the new name so the DAG keys are unique; verify
imports/strings and run tests/linters before merging.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In dependency_analysis.py around lines 312 to 324, the function

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033136

- [review_comment] 2025-09-18T15:55:19Z by coderabbitai[bot] (dependency_analysis.py:324)

```text
In dependency_analysis.py around lines 312 to 324, the function
get_normalized_feature_map() currently declares a return type of dict[str,
dict[str, list[str]]], but the payload includes original_name: str (a plain
string) causing a type mismatch; update the typing to accurately reflect the
shape (preferably define a TypedDict like FeatureNormalized { original_name:
str; hard: list[str]; soft: list[str]; enables: list[str]; provides: list[str] }
and change the function annotation to dict[str, FeatureNormalized]) and keep the
payload as-is (ensuring provides is a list[str]) or alternatively change
original_name to a single-element list if you prefer lists-only—apply one of
these fixes and update any imports/aliases accordingly.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/admin-api/docker-compose.yaml around lines 35-36, the config

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033150

- [review_comment] 2025-09-18T15:55:19Z by coderabbitai[bot] (deployments/admin-api/docker-compose.yaml:36)

```text
In deployments/admin-api/docker-compose.yaml around lines 35-36, the config
volume is mounted to /root/configs but the app reads from /app/config (defaults
config/config.yaml and config/admin-api.yaml); change the mount to map your
local config directory into /app/config (for example ./config:/app/config:ro),
keep the audit-logs volume as-is, and ensure the ./config host directory exists
and contains the expected config/config.yaml and config/admin-api.yaml files.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/admin-api/k8s-deployment.yaml around lines 41 to 44, the inline

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033161

- [review_comment] 2025-09-18T15:55:19Z by coderabbitai[bot] (deployments/admin-api/k8s-deployment.yaml:44)

```text
In deployments/admin-api/k8s-deployment.yaml around lines 41 to 44, the inline
comments for jwt-secret and redis-password lack the required two spaces before
the comment delimiter; update those lines to insert two spaces before each
inline comment so they read with a double-space separator before the "#" (e.g.,
value then two spaces then "# comment"), preserving the existing comment text
and alignment.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/docker/admin-api.env.example lines 1-4, replace the toy token

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033178

- [review_comment] 2025-09-18T15:55:19Z by coderabbitai[bot] (deployments/docker/admin-api.env.example:4)

```text
In deployments/docker/admin-api.env.example lines 1-4, replace the toy token
placeholders with clear secure-generation guidance and aligned names: require
tokens to be >=32 bytes entropy (provide examples for base64 and hex), mandate
role-prefixed keys (e.g. rq_admin_..., rq_read_), include example generation
commands (openssl rand -base64 32 and openssl rand -hex 32), instruct rotatation
and secure storage, and update placeholders to non-trivial examples; ensure the
env var names (API_TOKEN_1/API_TOKEN_2) and the secret key names
(api-token-1/api-token-2) remain consistent with
deployments/kubernetes/admin-api-deployment.yaml and deployments/README.md.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/docker/Dockerfile.admin-api around lines 52 to 53, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033189

- [review_comment] 2025-09-18T15:55:19Z by coderabbitai[bot] (deployments/docker/Dockerfile.admin-api:53)

```text
In deployments/docker/Dockerfile.admin-api around lines 52 to 53, the
HEALTHCHECK is targeting /health while other artifacts use /healthz; update the
HEALTHCHECK command to use the canonical /healthz path instead of /health (i.e.,
replace the URL to http://localhost:8080/healthz) so Dockerfile, Compose and K8s
probes are consistent and keep the existing curl flags and timing options
unchanged.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 9 | Runtime layer now installs curl and the healthcheck aligns with other images. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Runtime layer now installs curl and the healthcheck aligns with other images. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_005.md:144
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/kubernetes/admin-api-deployment.yaml around lines 114 to 123 (and

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033200

- [review_comment] 2025-09-18T15:55:19Z by coderabbitai[bot] (deployments/kubernetes/admin-api-deployment.yaml:123)

```text
In deployments/kubernetes/admin-api-deployment.yaml around lines 114 to 123 (and
also lines 125-126), currently API tokens are injected as environment variables
which leaks secrets; replace those env entries with a projected/secret volume:
define a volume that sources the admin-api-secrets secret, mount it into the
container at /var/run/secrets/admin-api, remove the API_TOKEN_1/API_TOKEN_2 env
entries, and update the app startup/config to read tokens/passwords from the
files /var/run/secrets/admin-api/api-token-1 and
/var/run/secrets/admin-api/api-token-2 instead of from environment variables.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/kubernetes/admin-api-deployment.yaml around lines 134 to 147, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033213

- [review_comment] 2025-09-18T15:55:19Z by coderabbitai[bot] (deployments/kubernetes/admin-api-deployment.yaml:147)

```text
In deployments/kubernetes/admin-api-deployment.yaml around lines 134 to 147, the
liveness probe path uses /health while the standard is /healthz and readiness
should be /readyz; update the liveness httpGet.path to /healthz, confirm
readiness remains /readyz, and ensure both probe port and timing settings remain
unchanged; then search and align Dockerfile/Compose and other deployment
artifacts to use /healthz and /readyz consistently.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## deployments/scripts/lib/logging.sh lines 4-6: the guard pattern that

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033228

- [review_comment] 2025-09-18T15:55:20Z by coderabbitai[bot] (deployments/scripts/lib/logging.sh:6)

```text
deployments/scripts/lib/logging.sh lines 4-6: the guard pattern that
returns/exists when the script is already sourced triggers shellcheck SC2317;
annotate it to avoid accidental future changes. Add a ShellCheck directive
immediately above the guard (e.g. a comment disabling SC2317) so the intentional
return/exit is documented and not flagged, and include a brief comment
explaining why the guard is needed; do not change the guard logic itself.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/scripts/test-staging-deployment.sh around lines 278 to 287,

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033256

- [review_comment] 2025-09-18T15:55:20Z by coderabbitai[bot] (deployments/scripts/test-staging-deployment.sh:287)

```text
In deployments/scripts/test-staging-deployment.sh around lines 278 to 287,
remove the blind sleep and instead poll the TCP socket (or HTTP endpoint) until
it becomes available, honoring a TIMEOUT environment variable to avoid hangs;
implement a loop that repeatedly attempts to connect (e.g., with curl -sSf or a
simple /dev/tcp check or nc) with short sleeps between tries and aborts with a
non-zero exit if the timeout is reached, then proceed to the health endpoint
test only after the socket/HTTP check succeeds.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/scripts/test-staging-deployment.sh around lines 336 to 346,

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033270

- [review_comment] 2025-09-18T15:55:20Z by coderabbitai[bot] (deployments/scripts/test-staging-deployment.sh:346)

```text
In deployments/scripts/test-staging-deployment.sh around lines 336 to 346,
replace the blind "sleep 5" before fetching the bootstrap token with the same
readiness loop used in the RBAC tests: poll kubectl (with a timeout and short
sleep interval) until the rbac-secrets secret (or its admin-bootstrap-token
field) exists and is readable, then proceed to read and base64-decode the token;
ensure the loop exits with a clear error if the secret never appears within the
timeout so the script fails fast instead of waiting blindly.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/_index.md around lines 7–13, the versioning policy is vague; update

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033294

- [review_comment] 2025-09-18T15:55:20Z by coderabbitai[bot] (docs/api/_index.md:13)

```text
In docs/api/_index.md around lines 7–13, the versioning policy is vague; update
it to explicitly require the following: list required HTTP headers and semantics
— a Sunset header with an absolute RFC‑1123 timestamp, Link headers with
rel="sunset" and rel="deprecation" pointing to the deprecation/remove notices,
and an optional Deprecation header containing version/date; define explicit
LTS/support windows per major (e.g., state “Each major is maintained for 18
months after the next major GA” or replace with your org’s chosen N months),
require that deprecated endpoints include an explicit removal date in both the
API docs and error response bodies once past deprecation, and mandate that all
path examples across docs use the versioned prefix /api/v1/... so route examples
are consistent.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/calendar-view.md around lines 84 to 101, the example and any

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033300

- [review_comment] 2025-09-18T15:55:21Z by coderabbitai[bot] (docs/api/calendar-view.md:101)

```text
In docs/api/calendar-view.md around lines 84 to 101, the example and any
documented endpoints omit the required API version prefix; update all path
references and example requests to use the /api/v1 prefix (e.g., change
/calendar/data to /api/v1/calendar/data) and apply the same change to all other
documented endpoints mentioned (/events, /reschedule, /rules, /config,
/timezones, /health, /debug/*) so every path in this file consistently uses
/api/v1.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/calendar-view.md around lines 465 to 491, the example response leaks

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033314

- [review_comment] 2025-09-18T15:55:21Z by coderabbitai[bot] (docs/api/calendar-view.md:491)

```text
In docs/api/calendar-view.md around lines 465 to 491, the example response leaks
numeric enum values (default_view: 0 and action numeric codes) which are
client‑hostile; update the JSON examples to emit enum names as strings (e.g.,
"default_view": "<EnumName>" and each "action": "<ActionName>") and update any
surrounding text to state that the server accepts and returns string enum names
(while noting the Go SDK may map those names to ints internally). Ensure all
key_bindings.action fields and default_view use the string names throughout the
example and add a short note clarifying server behavior and the Go SDK mapping.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/calendar-view.md around lines 519 to 548, the error examples and

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033324

- [review_comment] 2025-09-18T15:55:21Z by coderabbitai[bot] (docs/api/calendar-view.md:548)

```text
In docs/api/calendar-view.md around lines 519 to 548, the error examples and
table mix numeric "code" values with string-style error identifiers; update the
documentation to use stable string error codes everywhere (e.g.
"ErrorCodeEventNotFound") instead of numeric codes in examples and the table,
keep numeric codes as internal implementation details only, and add a note/link
to a new docs/error_codes.md that lists the numeric->string mapping for humans;
ensure the JSON example uses "error_code" (string) consistently, update any
schema/response examples in this section to reflect the "error_code" field and
remove the numeric "code" field, and mention that services should still
translate to numeric codes internally but expose string codes in the public API
docs.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/calendar-view.md around lines 739 to 758, the authentication

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033336

- [review_comment] 2025-09-18T15:55:21Z by coderabbitai[bot] (docs/api/calendar-view.md:758)

```text
In docs/api/calendar-view.md around lines 739 to 758, the authentication
instructions mix an explicit X-User-ID header with JWT-based identity; clarify
that identity is derived from validated JWT claims and either remove the
X-User-ID example or mark it as internal/testing-only and explicitly ignored
when a valid Authorization: Bearer <jwt> is provided; if you choose to support
both, document that X-User-ID must be validated against the JWT sub/claims on
the server and only accepted when it matches after strict verification.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/SLAPS/coordinator-observations.md around lines 195 to 197, fix the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033350

- [review_comment] 2025-09-18T15:55:21Z by coderabbitai[bot] (docs/SLAPS/coordinator-observations.md:197)

```text
In docs/SLAPS/coordinator-observations.md around lines 195 to 197, fix the
grammar in the first bullet by changing "a entire microservice architecture in
parallel" to "an entire microservice architecture in parallel"; update only that
article ("a" → "an") to match correct English usage.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/SLAPS/coordinator-observations.md around lines 235 to 236, the "Total

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033379

- [review_comment] 2025-09-18T15:55:21Z by coderabbitai[bot] (docs/SLAPS/coordinator-observations.md:236)

```text
In docs/SLAPS/coordinator-observations.md around lines 235 to 236, the "Total
Runtime: ~7 hours (with two 4.5-hour rate limit pauses)" phrasing is confusing;
change it to clearly separate active compute time and wall-clock time by
replacing that line with something like "Active compute time: ~7 hours;
Wall-clock time (including two 4.5-hour pauses): ~16 hours" so readers
immediately understand the distinction between processing time and elapsed time.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/SLAPS/worker-reflections/claude-001-reflection.md around lines 39 to 41

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033393

- [review_comment] 2025-09-18T15:55:21Z by coderabbitai[bot] (docs/SLAPS/worker-reflections/claude-001-reflection.md:41)

```text
In docs/SLAPS/worker-reflections/claude-001-reflection.md around lines 39 to 41
the docs claim a METRICS_ENABLED toggle and per-test registry but the codebase
lacks that guard and still uses global prometheus.MustRegister calls; update
internal/obs/metrics.go (init around line ~66) and
internal/producer-backpressure/types.go (BackpressureMetrics.Register around
line ~291) to remove package-global MustRegister usage and instead accept an
injected prometheus.Registerer or gate registrations behind a configuration flag
(METRICS_ENABLED / observability.metrics.enabled). Implement one of the two
options: A) add a config-driven guard that skips global registration and allow
tests to provide prometheus.NewRegistry(), or B) refactor so no init()
registrations occur and all Register methods take a Registerer parameter (use
dependency injection), then update tests to inject per-test registries and add a
regression test to assert no duplicate-collector panics occur.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/YOU ARE WORKER 6.md around lines 8–16, the current claim protocol is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033401

- [review_comment] 2025-09-18T15:55:22Z by coderabbitai[bot] (docs/YOU ARE WORKER 6.md:16)

```text
In docs/YOU ARE WORKER 6.md around lines 8–16, the current claim protocol is
racy across filesystems; replace the mv-based approach with an atomic claim
procedure that stages the file on the target filesystem and performs an atomic
rename or uses an O_CREAT|O_EXCL lock to fail if another worker already claimed
it. Specifically: create a temp file in the worker directory (so it lives on the
same FS as the destination), copy the source into that temp, attempt an atomic
exclusive claim (e.g., create/link a lockname using O_CREAT|O_EXCL or ln to fail
if lock exists), on success rename the temp to the final target atomically,
remove the lock and then remove the original source only after verifying the
rename succeeded, and on any failure leave the source untouched and log the
error.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 248 to 255, the perf table

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033421

- [review_comment] 2025-09-18T15:55:22Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:255)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 248 to 255, the perf table
incorrectly states Go 1.22.5 while repo go.mod files declare Go 1.25/1.25.0;
update the table to list Go 1.25 (or re-run the benchmarks under Go 1.22.5 and
replace the benchmark numbers if you prefer to keep 1.22.5), and add a note
about the exact Go toolchain used (including patch version) and where raw
outputs live; also ensure CI/workflows pin the Go version used for benchmarking
(update .github workflows to use go-version: 1.25.x) so results are
reproducible.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 318 to 321, the claim that

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360033449

- [review_comment] 2025-09-18T15:55:22Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:321)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 318 to 321, the claim that
"All tests are deterministic with controlled randomness" needs concrete details:
update the section to state that all randomized tests accept a TEST_SEED
environment variable and show example CI/test commands including the shuffle
flag (e.g., add text like "All randomized tests accept TEST_SEED; CI runs with
-shuffle=on" and an example command showing -shuffle=on), and ensure you mention
where to set TEST_SEED locally and in CI so readers can reproduce deterministic
runs.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In .github/workflows/update-progress.yml around lines 30 to 33, the workflow

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360038808

- [review_comment] 2025-09-18T15:56:31Z by coderabbitai[bot] (.github/workflows/update-progress.yml:33)

```text
In .github/workflows/update-progress.yml around lines 30 to 33, the workflow
currently runs python3 scripts/update_progress.py without verifying the script
exists; add a pre-check that verifies scripts/update_progress.py is present and,
if not, echoes a clear error and exits non-zero so the job fails fast instead of
attempting to run a non-existent script, then only invoke python3 when the file
check passes.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In create_review_tasks.py around lines 107-112 (and also line 141), the test

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360038835

- [review_comment] 2025-09-18T15:56:31Z by coderabbitai[bot] (create_review_tasks.py:112)

```text
In create_review_tasks.py around lines 107-112 (and also line 141), the test
coverage threshold is inconsistent (90% in one place vs 80% elsewhere); pick a
single canonical threshold (e.g., 90%) and update every occurrence in this file
to match it so docs and checks agree — search for any "80%" or "90%" coverage
strings or numeric threshold variables in the file and replace them with the
chosen value, and ensure any related comments/messages reflect the same
threshold.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/admin-api/k8s-deployment.yaml around lines 71 to 116, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360038848

- [review_comment] 2025-09-18T15:56:31Z by coderabbitai[bot] (deployments/admin-api/k8s-deployment.yaml:116)

```text
In deployments/admin-api/k8s-deployment.yaml around lines 71 to 116, the
JWT_SECRET and REDIS_PASSWORD are exposed as environment variables; change to
mount the admin-api-secrets Secret as a volume and remove those secret ENV
entries, mounting the secret files (e.g. /etc/secrets/jwt-secret and
/etc/secrets/redis-password) into the container with appropriate readOnly
permissions and update volume/volumeMounts to reference the Secret; then update
the application startup logic to first read credentials from the mounted files
if present (fallback to env only if files missing) and ensure file permissions
and container user (runAsUser 1000) can read them.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Hardened the admin API deployment with pod/container security contexts and RuntimeDefault seccomp. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Hardened the admin API deployment with pod/container security contexts and RuntimeDefault seccomp. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_010.md:596
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/admin-api/k8s-deployment.yaml around lines 152–167, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360038867

- [review_comment] 2025-09-18T15:56:32Z by coderabbitai[bot] (deployments/admin-api/k8s-deployment.yaml:167)

```text
In deployments/admin-api/k8s-deployment.yaml around lines 152–167, the
annotation nginx.ingress.kubernetes.io/rate-limit is non‑standard and will be
ignored; replace it with a documented ingress‑nginx annotation such as
nginx.ingress.kubernetes.io/limit-rps: "100" (or
nginx.ingress.kubernetes.io/limit-rpm: "6000" if you prefer per‑minute limits)
and remove any non‑standard keys like
nginx.ingress.kubernetes.io/rate-limit-window (or map it to an appropriate
documented setting if needed), then redeploy and validate that the controller
enforces the configured limits; apply the same replacement for the other files
referenced in the review.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/docker/grafana/datasources/prometheus.yaml around lines 3 to 9,

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360038885

- [review_comment] 2025-09-18T15:56:32Z by coderabbitai[bot] (deployments/docker/grafana/datasources/prometheus.yaml:9)

```text
In deployments/docker/grafana/datasources/prometheus.yaml around lines 3 to 9,
the Prometheus datasource is missing a uid so dashboards that reference uid
"Prometheus" will 404; add a fixed uid field (set uid: Prometheus to match the
dashboards) to the datasource definition so it can be reliably referenced, keep
the rest of the fields unchanged.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/kubernetes/rbac-monitoring.yaml around lines 35 to 42, the alert

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360038904

- [review_comment] 2025-09-18T15:56:32Z by coderabbitai[bot] (deployments/kubernetes/rbac-monitoring.yaml:42)

```text
In deployments/kubernetes/rbac-monitoring.yaml around lines 35 to 42, the alert
expr uses up{app="rbac-token-service"} == 0 which misses the case where all
targets disappear; replace the expr with a combined check using sum and absent,
e.g. use an expression that evaluates true when either the summed up is zero or
the series is absent (for example: sum(up{app="rbac-token-service"}) == 0 or
absent(up{app="rbac-token-service"})), leaving the for, labels and annotations
unchanged.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/kubernetes/rbac-monitoring.yaml around lines 45 to 51 (and also

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360038916

- [review_comment] 2025-09-18T15:56:32Z by coderabbitai[bot] (deployments/kubernetes/rbac-monitoring.yaml:51)

```text
In deployments/kubernetes/rbac-monitoring.yaml around lines 45 to 51 (and also
lines 88 to 100), the alert uses job="rbac-token-service" while other
rules/dashboards use app="rbac-token-service", causing alerts to miss metrics;
pick one label (recommended: app) and update the PromQL selectors to the chosen
label consistently (e.g., replace job="rbac-token-service" with
app="rbac-token-service" in this alert and the other affected rules), and verify
any relabeling rules export that label so both metrics and alerts match.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/kubernetes/rbac-monitoring.yaml around lines 161–173, the two

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360038921

- [review_comment] 2025-09-18T15:56:32Z by coderabbitai[bot] (deployments/kubernetes/rbac-monitoring.yaml:173)

```text
In deployments/kubernetes/rbac-monitoring.yaml around lines 161–173, the two
alerts reference metrics (rbac_admin_actions_total and
rbac_key_last_rotation_timestamp) that are not exported by the RBAC service;
either implement and register those metrics in the RBAC service or
remove/replace these alerts with existing signals. To implement: add Prometheus
metric definitions (e.g., prometheus.NewCounter for admin actions and
prometheus.NewGauge or prometheus.NewGaugeFunc for last-rotation timestamp) in
the RBAC service code, register them with prometheus.MustRegister (see
internal/obs/metrics.go and internal/producer-backpressure/types.go for
examples), and update instrumentation to increment/set them; to remove/replace:
delete these alert blocks from rbac-monitoring.yaml or change expr to use valid
metrics already exported by the service.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/kubernetes/rbac-token-service-deployment.yaml around lines 367 to

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360038937

- [review_comment] 2025-09-18T15:56:32Z by coderabbitai[bot] (deployments/kubernetes/rbac-token-service-deployment.yaml:371)

```text
In deployments/kubernetes/rbac-token-service-deployment.yaml around lines 367 to
371, the annotations using nginx.ingress.kubernetes.io/rate-limit* are invalid;
replace them with the real NGINX ingress annotations
nginx.ingress.kubernetes.io/limit-rps and
nginx.ingress.kubernetes.io/limit-burst and set values equivalent to the
previous intent (e.g., for 60 requests per minute use
nginx.ingress.kubernetes.io/limit-rps: "1" and
nginx.ingress.kubernetes.io/limit-burst: "60"), keeping the other annotations
(rewrite-target, ssl-redirect, cert-manager) unchanged.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/scripts/deploy-rbac-staging.sh around line 11, ShellCheck is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360038953

- [review_comment] 2025-09-18T15:56:33Z by coderabbitai[bot] (deployments/scripts/deploy-rbac-staging.sh:11)

```text
In deployments/scripts/deploy-rbac-staging.sh around line 11, ShellCheck is
warning about sourcing a local library; add a ShellCheck source directive
immediately above the source line to point to the lib path (for example: "#
shellcheck source=lib/logging.sh") so ShellCheck knows where the sourced file
lives, then keep the existing source "${SCRIPT_DIR}/lib/logging.sh" line
unchanged.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/scripts/deploy-rbac-staging.sh around lines 13 to 35, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360038970

- [review_comment] 2025-09-18T15:56:33Z by coderabbitai[bot] (deployments/scripts/deploy-rbac-staging.sh:35)

```text
In deployments/scripts/deploy-rbac-staging.sh around lines 13 to 35, the
prerequisite check function currently verifies kubectl and docker but misses
validating required tools openssl and curl; add checks similar to the existing
ones: test command -v openssl and command -v curl, emit an error message and
exit 1 if either is missing, and include them before the cluster connect check
so the script fails fast with clear guidance to install the missing utilities.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/scripts/deploy-rbac-staging.sh around lines 36 to 47, the build

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360038975

- [review_comment] 2025-09-18T15:56:33Z by coderabbitai[bot] (deployments/scripts/deploy-rbac-staging.sh:47)

```text
In deployments/scripts/deploy-rbac-staging.sh around lines 36 to 47, the build
step creates and tags a local image work-queue/rbac-token-service:staging which
never gets pushed and does not match the manifest's pinned image; add
IMAGE_REPO, IMAGE_TAG and IMAGE variables near the top (after line 7) to match
the Deployment, update build_image to build with -t "$IMAGE", push the image to
the registry (docker push "$IMAGE"), and remove or stop using the local-only
staging tag so the deployed manifest pulls the exact pushed tag.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/scripts/deploy-rbac-staging.sh around lines 65 to 68, the script

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360038984

- [review_comment] 2025-09-18T15:56:33Z by coderabbitai[bot] (deployments/scripts/deploy-rbac-staging.sh:68)

```text
In deployments/scripts/deploy-rbac-staging.sh around lines 65 to 68, the script
currently uses "export RBAC_SIGNING_KEY RBAC_ENCRYPTION_KEY REDIS_PASSWORD
ADMIN_BOOTSTRAP_TOKEN" which unnecessarily places secrets into the environment
and risks leaking them; change these to plain shell variables (assign them
without export) so they remain in-script only (e.g., RBAC_SIGNING_KEY="..."
RBAC_ENCRYPTION_KEY="..." REDIS_PASSWORD="..." ADMIN_BOOTSTRAP_TOKEN="..."),
remove the export statement, and ensure no subsequent commands rely on these
variables being inherited by child processes; optionally unset the variables
before script exit for extra safety.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/scripts/deploy-staging.sh around lines 184-185 you call

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039024

- [review_comment] 2025-09-18T15:56:34Z by coderabbitai[bot] (deployments/scripts/deploy-staging.sh:185)

```text
In deployments/scripts/deploy-staging.sh around lines 184-185 you call
register_port_forward "$PF_PID" twice which double-registers the same PID and
will attempt to kill it twice; remove the duplicated register_port_forward
invocation so the PID is registered only once (leave a single
register_port_forward "$PF_PID" call) and optionally add a sanity check that
PF_PID is non-empty before calling to avoid registering an empty value.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/scripts/setup-monitoring.sh around line 11, the script blindly

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039038

- [review_comment] 2025-09-18T15:56:34Z by coderabbitai[bot] (deployments/scripts/setup-monitoring.sh:11)

```text
In deployments/scripts/setup-monitoring.sh around line 11, the script blindly
sources "${SCRIPT_DIR}/lib/logging.sh" which may not exist; add a guard that
checks the file is present and readable before sourcing, and if missing print a
clear error to stderr and exit non‑zero (fail fast). Use a conditional to test
-r or -f on "${SCRIPT_DIR}/lib/logging.sh" and only source it when the check
passes; otherwise echo a descriptive error to >&2 and exit 1 so ShellCheck
SC1091 is addressed.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/scripts/setup-monitoring.sh around lines 103 to 108, the scrape

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039054

- [review_comment] 2025-09-18T15:56:34Z by coderabbitai[bot] (deployments/scripts/setup-monitoring.sh:108)

```text
In deployments/scripts/setup-monitoring.sh around lines 103 to 108, the scrape
config sets honorLabels: true which allows targets to override job/instance
labels; remove this line or set honorLabels: false to prevent targets from
clobbering labels and breaking grouping, then re-generate/validate the resulting
Prometheus config and restart/reload the monitoring stack to apply the change.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/scripts/setup-monitoring.sh around lines 146 to 218, the script

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039058

- [review_comment] 2025-09-18T15:56:34Z by coderabbitai[bot] (deployments/scripts/setup-monitoring.sh:218)

```text
In deployments/scripts/setup-monitoring.sh around lines 146 to 218, the script
checks for the secret name "alertmanager-main" but creates
"alertmanager-rbac-config", so detection and creation refer to different
resources; make them consistent by changing the kubectl get secret check to look
for "alertmanager-rbac-config" (and update any log messages if needed), or
alternatively change the created secret name to "alertmanager-main" so both
lines reference the same secret; modify only the secret name in the detection or
creation block to match the other and keep log messages aligned.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/scripts/setup-monitoring.sh around lines 211 to 215, the secret

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039081

- [review_comment] 2025-09-18T15:56:35Z by coderabbitai[bot] (deployments/scripts/setup-monitoring.sh:215)

```text
In deployments/scripts/setup-monitoring.sh around lines 211 to 215, the secret
is created using --from-literal which leaks the config into process arguments;
instead feed the config via stdin or --from-file reading from /dev/stdin.
Replace the --from-literal usage with a method that pipes the
$alertmanager_config into kubectl (for example using a here-doc or printf |
kubectl and --from-file=alertmanager.yml=/dev/stdin or by constructing the
secret YAML on stdin) so the secret contents do not appear in the process list
or shell args; ensure quoting/encoding is preserved when piping and then
continue to dry-run -o yaml | kubectl apply -f - as before.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/07_test_plan.md around lines 53 to 58, the CI/workflow and docs are out

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039120

- [review_comment] 2025-09-18T15:56:35Z by coderabbitai[bot] (docs/07_test_plan.md:58)

```text
In docs/07_test_plan.md around lines 53 to 58, the CI/workflow and docs are out
of sync: workflows don't set GOMAXPROCS or BENCHMARK_SEED and the runner is not
pinned to the documented ubuntu-22.04/8vCPU instance; update the CI benchmarking
job to export BENCHMARK_SEED (and log it with results) and set GOMAXPROCS=8 (via
job env or export before running benchmarks), and either pin runs-on to the
documented ubuntu-22.04 runner type (or the exact instance type) in
.github/workflows/* where benchmarks run or change this doc line to match the
actual runner used; also ensure the synthetic producer accepts a seed parameter
and that the workflows record the chosen seed in the test artifacts/logs.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/12_performance_baseline.md around lines 46 to 53, the instruction to set

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039143

- [review_comment] 2025-09-18T15:56:36Z by coderabbitai[bot] (docs/12_performance_baseline.md:53)

```text
In docs/12_performance_baseline.md around lines 46 to 53, the instruction to set
worker.count to 16 on a 4 vCPU node is a magic number; change it to recommend
calculating worker.count as k × GOMAXPROCS (or runtime.GOMAXPROCS(0)) and
provide guidance for choosing k (e.g., k≈1 for CPU-bound workloads, k=2–4+ for
I/O-bound or network-heavy workloads), include how to detect workload type and
an example calculation (e.g., with 4 GOMAXPROCS and k=2 → worker.count=8), and
mention that users should tune based on measured latency/throughput and system
limits rather than using a fixed 16.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/anomaly-radar-openapi.yaml around lines 80-109 (and similarly

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039155

- [review_comment] 2025-09-18T15:56:36Z by coderabbitai[bot] (docs/api/anomaly-radar-openapi.yaml:109)

```text
In docs/api/anomaly-radar-openapi.yaml around lines 80-109 (and similarly
404-421), several array schemas lack maxItems which can lead to unbounded
responses; add a maxItems property to every array-type schema and array query
parameter (e.g., the alerts array in AlertsResponse and validation_errors in
ValidationErrorResponse) and ensure their limits make sense for the field
(suggest setting alerts maxItems to a sane upper bound like 1000,
validation_errors to something smaller like 100, and align any snapshot list
maxItems with the max_samples cap), updating any related descriptions to reflect
the cap.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/anomaly-radar-openapi.yaml around lines 229 to 236 (and similarly

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039164

- [review_comment] 2025-09-18T15:56:36Z by coderabbitai[bot] (docs/api/anomaly-radar-openapi.yaml:236)

```text
In docs/api/anomaly-radar-openapi.yaml around lines 229 to 236 (and similarly
254-266), replace any inline short/brace map usage like "{ key: value }" with
expanded standard YAML block mappings: put each key on its own line under the
parent with proper indentation and no braces or extra spaces inside braces; do
this for the securitySchemes/description and for all response blocks that
currently use inline brace maps so they conform to yamllint rules about spacing
and mapping style.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/anomaly-radar-openapi.yaml around lines 386 to 393, the description

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039169

- [review_comment] 2025-09-18T15:56:36Z by coderabbitai[bot] (docs/api/anomaly-radar-openapi.yaml:393)

```text
In docs/api/anomaly-radar-openapi.yaml around lines 386 to 393, the description
for the Go time.ParseDuration example currently shows a shorthand (e.g., 720h)
but our docs use canonical duration strings elsewhere; update the description to
include the canonical format (e.g., 720h0m0s) and, if appropriate, show both
short and canonical examples (e.g., "Go time.ParseDuration format (e.g., 720h or
720h0m0s)"), ensuring wording and examples match the project's canonical
duration style used in other docs.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/anomaly-radar-openapi.yaml around lines 433-446, numeric fields lack

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039176

- [review_comment] 2025-09-18T15:56:36Z by coderabbitai[bot] (docs/api/anomaly-radar-openapi.yaml:446)

```text
In docs/api/anomaly-radar-openapi.yaml around lines 433-446, numeric fields lack
bounds; add validation constraints: set error_rate to minimum: 0 and maximum: 1;
set error_count to minimum: 0 (integer); set p50_latency_ms, p90_latency_ms,
p95_latency_ms, p99_latency_ms to minimum: 0 (number); and apply equivalent
min/max rules for any other percentile or threshold fields elsewhere (percentile
probabilities use [0,1], counts use >=0, latencies/thresholds use >=0 or
appropriate upper limits).
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## docs/api/anomaly-radar-slo-budget.md around lines 223 to 241: the /status

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039186

- [review_comment] 2025-09-18T15:56:37Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:241)

```text
docs/api/anomaly-radar-slo-budget.md around lines 223 to 241: the /status
payload currently exposes "config" at the top-level of "slo_budget" while
/config uses a nested shape under "slo" -> "thresholds", which will break
clients; either remove "config" from /status or, preferably, change the /status
example to match /config by nesting those fields under "slo": { "thresholds": {
... } } (preserve the same keys and values), update any surrounding text to
reference slo.thresholds instead of slo_budget.config, and ensure timestamp and
other fields remain at the same levels.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## docs/api/canary-deployments.md around lines 68-76 (and also apply to 366-373,

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039200

- [review_comment] 2025-09-18T15:56:37Z by coderabbitai[bot] (docs/api/canary-deployments.md:76)

```text
docs/api/canary-deployments.md around lines 68-76 (and also apply to 366-373,
375-380, 783-789): the example JSON uses IDs like "canary_<uuid>" while the spec
later mandates ULIDs; choose one consistent ID scheme and update both examples
and the spec. Either (A) change examples to plain ULIDs (remove the "canary_"
prefix) and update any example values/show regex to match ULID format, or (B)
update the spec to state "prefix + ULID", adjust the descriptive text and
provide a regex that matches the "canary_" prefix followed by a ULID, then
update all listed example occurrences to follow that chosen pattern so examples
and regexes are consistent.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/canary-deployments.md around lines 84-90 (and likewise at 112-118,

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039213

- [review_comment] 2025-09-18T15:56:37Z by coderabbitai[bot] (docs/api/canary-deployments.md:90)

```text
In docs/api/canary-deployments.md around lines 84-90 (and likewise at 112-118,
120-131, 508-519, 527-553, 579-589), the duration fields and their example
values are inconsistent (mixing max_duration/min_duration with
max_canary_duration/min_canary_duration and non-canonical formats). Standardize
to a single canonical schema (use max_canary_duration and min_canary_duration
everywhere), convert all example duration values to Go time.Duration canonical
strings (e.g., "2h0m0s", "5m0s"), and update the Parameters section text and
each profile example to reference the exact field names and formats
consistently.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/canary-deployments.md around lines 90-92 (and also update lines

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039224

- [review_comment] 2025-09-18T15:56:37Z by coderabbitai[bot] (docs/api/canary-deployments.md:92)

```text
In docs/api/canary-deployments.md around lines 90-92 (and also update lines
128-131 and 156-164), the spec lacks numeric validation for percentage updates;
add a clear constraint that percentage values must be numeric between 0 and 100
inclusive and may include decimals up to 2 decimal places, and specify that
requests with values outside this range or invalid formats must return HTTP 400;
update the JSON schema/examples and plain-language description in those sections
to state the allowed range, decimal precision, and the 400 error response for
invalid input.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/canary-deployments.md around lines 90-92 (and also adjust at lines

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039224

- [review_comment] 2025-09-18T15:56:37Z by coderabbitai[bot] (docs/api/canary-deployments.md:92)

```text
In docs/api/canary-deployments.md around lines 90-92 (and also adjust at lines
118 and 744-756), the response envelope is inconsistent: it returns
"deployments":[...], "count": n while the rest of the docs use data/pagination.
Update the examples to a consistent envelope that uses "data" with a
"pagination" object (e.g. wrap lists under data.<resource> and replace top-level
"count" with data.pagination { total, limit, offset/page }), and apply the same
change to Events and Workers list response examples so all list endpoints use
the same data/pagination structure.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/canary-deployments.md around lines 304-316 (and also adjust

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039243

- [review_comment] 2025-09-18T15:56:37Z by coderabbitai[bot] (docs/api/canary-deployments.md:316)

```text
In docs/api/canary-deployments.md around lines 304-316 (and also adjust
occurrences at 608-610 and 626), the metrics snapshot percent fields are
ambiguous; ensure all percent fields use a 0–100 percentage scale (not
fractions) and add a single clarifying sentence to the Metrics Snapshot section
stating "All percent fields (error_percent, success_percent, etc.) are expressed
on a 0–100 scale (e.g., 0.96 means 0.96%)." Update the example values and any
nearby percent descriptions to match that convention and verify the other
referenced lines (608-610, 626) use the same wording and numeric format.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/canary-deployments.md around lines 732-739 (and similarly lines

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039258

- [review_comment] 2025-09-18T15:56:37Z by coderabbitai[bot] (docs/api/canary-deployments.md:739)

```text
In docs/api/canary-deployments.md around lines 732-739 (and similarly lines
13-21), clarify the ambiguous "per API key" rate limit by explicitly stating the
exact subject used for counting: whether it's the raw API token string for token
auth, or the JWT's subject claim (e.g., `sub`) or tenant identifier for JWT
auth; state both cases if both auth methods are supported. Update the rate limit
bullets to name the exact key used for each auth method, confirm that the
X-RateLimit-* headers are emitted and identical for both authentication methods,
and mirror the same precise language in the Authentication section so both
places describe the same subject and header behavior.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/capacity-planning-api.md around lines 317 to 324, the example

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039292

- [review_comment] 2025-09-18T15:56:38Z by coderabbitai[bot] (docs/api/capacity-planning-api.md:324)

```text
In docs/api/capacity-planning-api.md around lines 317 to 324, the example
instructs importing an internal package which will fail for external consumers;
either move the package out of internal into pkg/ with a stable public import
path and update the example import to that new module path (and update
go.mod/tests/CI references accordingly), or explicitly mark this code block as
internal-only and remove/replace the import with the public, exported API you
want external users to consume; update the documentation to show the correct
public path and a short note about internal-only visibility.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/TUI/README.md around lines 31 to 33, replace the nonstandard admonition

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360039353

- [review_comment] 2025-09-18T15:56:38Z by coderabbitai[bot] (docs/TUI/README.md:33)

```text
In docs/TUI/README.md around lines 31 to 33, replace the nonstandard admonition
marker `[!note]-` with a GitHub-friendly pattern: either a simple blockquote
starting with "Note:" (e.g. `> **Note:** ...`) or a section header like `####
Note` (or remove the admonition entirely); apply the same change to all other
occurrences in this file so linters/renderers no longer fail on the `[!note]-`
syntax.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In .github/workflows/update-progress.yml around lines 34 to 54, the commit step

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360044726

- [review_comment] 2025-09-18T15:57:45Z by coderabbitai[bot] (.github/workflows/update-progress.yml:54)

```text
In .github/workflows/update-progress.yml around lines 34 to 54, the commit step
should explicitly mark the repo safe to avoid “dubious ownership” errors and
must handle empty staging reliably; add a git config --global --add
safe.directory "$(pwd)" (or "$GITHUB_WORKSPACE") before any git commands, ensure
the files_to_add array is only passed to git add when non-empty (as you already
guard) and replace the cached-diff check with a robust staged-change check such
as using git diff --cached --quiet || git diff --quiet to detect staged or
unstaged changes or use git diff --name-only --cached | grep -q . to determine
if there are staged files before committing, then only run git commit when there
is at least one staged change and set the changed output accordingly.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In .github/workflows/update-progress.yml around lines 56-58, the push step

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360044748

- [review_comment] 2025-09-18T15:57:46Z by coderabbitai[bot] (.github/workflows/update-progress.yml:58)

```text
In .github/workflows/update-progress.yml around lines 56-58, the push step
currently does a plain git push which will fail on non-fast-forward updates;
instead make the run step try to rebase the local changes onto the remote and
then push to avoid race failures — i.e., fetch the remote, perform a git pull
--rebase (or git rebase origin/<branch>) to incorporate upstream commits,
resolve/abort on conflicts as necessary, then git push; ensure the step still
only runs when changes exist.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/admin-api/Dockerfile around lines 18 to 21, the Go build step

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360044788

- [review_comment] 2025-09-18T15:57:46Z by coderabbitai[bot] (deployments/admin-api/Dockerfile:21)

```text
In deployments/admin-api/Dockerfile around lines 18 to 21, the Go build step
currently uses -trimpath but still embeds VCS metadata; to make builds
reproducible append -buildvcs=false to the ldflags so the -ldflags string
becomes "-s -w -X main.version=${VERSION} -buildvcs=false" (preserving existing
flags and output path) to disable VCS stamping.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/admin-api/Dockerfile around lines 19 to 21, the ldflags usage -X

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360044814

- [review_comment] 2025-09-18T15:57:46Z by coderabbitai[bot] (deployments/admin-api/Dockerfile:21)

```text
In deployments/admin-api/Dockerfile around lines 19 to 21, the ldflags usage -X
main.version=${VERSION} will fail because there is no var version string in
package main (cmd/admin-api); either add a top-level declaration in
cmd/admin-api (package main) like a var version string so -X main.version can
link, or change the Dockerfile ldflags to reference the exact exported variable
by its full package path and identifier (case‑sensitive), e.g. -X
'github.com/your/repo/cmd/admin-api.VarName=${VERSION}', and ensure proper
quoting/escaping in the Dockerfile.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## deployments/admin-api/Dockerfile lines 23-46: the image lacks standard OCI

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360044829

- [review_comment] 2025-09-18T15:57:46Z by coderabbitai[bot] (deployments/admin-api/Dockerfile:46)

```text
deployments/admin-api/Dockerfile lines 23-46: the image lacks standard OCI
metadata labels; add a LABEL instruction near the top of the Dockerfile
(immediately after the FROM) that sets common OCI labels such as
org.opencontainers.image.title, org.opencontainers.image.description,
org.opencontainers.image.version, org.opencontainers.image.revision (commit
SHA), org.opencontainers.image.created (build timestamp),
org.opencontainers.image.authors, org.opencontainers.image.licenses,
org.opencontainers.image.url/source and org.opencontainers.image.vendor;
implement values via build-time ARGs with sensible defaults (e.g., VERSION,
VCS_REF, BUILD_DATE, MAINTAINER) so CI can inject real values, and keep existing
functionality unchanged.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/admin-api/Dockerfile around lines 49 to 50, remove the Docker

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360044839

- [review_comment] 2025-09-18T15:57:46Z by coderabbitai[bot] (deployments/admin-api/Dockerfile:50)

```text
In deployments/admin-api/Dockerfile around lines 49 to 50, remove the Docker
HEALTHCHECK and the wget command so the image no longer includes a
container-side HTTP probe; delete the two HEALTHCHECK lines and any installation
of wget/alpine packages used solely for that check so the image is leaner and
rely on Kubernetes liveness/readiness probes instead.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In AGENTS.md around lines 791 to 828, there is a duplicated "## APPENDIX B: WILD

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360060955

- [review_comment] 2025-09-18T16:01:13Z by coderabbitai[bot] (AGENTS.md:828)

```text
In AGENTS.md around lines 791 to 828, there is a duplicated "## APPENDIX B: WILD
IDEAS — HAVE A BRAINSTORM" heading; remove the first/stray APPENDIX B block (the
one preceding the detailed NOTE and activity log) so only a single APPENDIX B
remains, and if you want to preserve the NOTE text, move that NOTE under the
Daily Activity Logs section instead of duplicating the appendix heading.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In append_metadata.py around lines 53 to 80, the function currently appends YAML

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360060966

- [review_comment] 2025-09-18T16:01:13Z by coderabbitai[bot] (append_metadata.py:80)

```text
In append_metadata.py around lines 53 to 80, the function currently appends YAML
front matter to the end of the file which breaks tooling; change the logic to
detect existing front matter at the top (use content.lstrip().startswith("---")
or content.startswith("---")) and, when no front matter exists, write the file
with yaml_metadata + "\n\n" + content.lstrip("\n") instead of appending at EOF;
keep the same read/try/except structure and update the checks and write call
accordingly so metadata is prepended with a blank line separator.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In append_metadata.py around lines 184-194, the DAG write is fine but the file

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360060975

- [review_comment] 2025-09-18T16:01:14Z by coderabbitai[bot] (append_metadata.py:194)

```text
In append_metadata.py around lines 184-194, the DAG write is fine but the file
has critical issues to fix elsewhere: remove the static infrastructure_nodes
list defined at/near line 87 and instead import infrastructure data from
dependency_analysis; replace the brittle front-matter check at/near line 67 (if
content.endswith("---")) with a proper YAML/front-matter parser that
extracts/loads the front-matter block (e.g., locate the leading/trailing '---'
and yaml.safe_load the slice) so edge cases are handled; stop hardcoding spec
paths at/near line 155 by constructing them with os.path.join(ideas_dir, "docs",
"ideas", f"{feature_name}.md") or the correct platform-safe path for your repo
layout; and add normalized imports at the top of the file: from
dependency_analysis import get_normalized_feature_map, normalize_name,
infrastructure (and any other required symbols) so the earlier removals use
those functions/values.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In BUGS.md around lines 1-7 (and similarly lines 131-136), the tone is informal

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360060981

- [review_comment] 2025-09-18T16:01:14Z by coderabbitai[bot] (BUGS.md:7)

```text
In BUGS.md around lines 1-7 (and similarly lines 131-136), the tone is informal
and uses slang (“slaps”, “no‑BS”, “SLAPS swarm”) which is inappropriate for
engineering documentation; replace slang with neutral, precise language, remove
hype and persona, and rewrite the paragraphs as concise engineering statements
that describe features and the punch list plainly (e.g., list implemented
features and required next steps), keep factual details (BRPOPLPUSH,
heartbeat/reaper, priorities, DLQ, tests, metrics, config) but present them in a
bulleted or numbered, professional format and ensure tone is consistent
throughout the file.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In BUGS.md around lines 14-21, the current code treats SetArgs as returning

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360060989

- [review_comment] 2025-09-18T16:01:14Z by coderabbitai[bot] (BUGS.md:21)

```text
In BUGS.md around lines 14-21, the current code treats SetArgs as returning
(bool, error) which is incorrect; change the logic to either use SetNX
(rdb.SetNX(ctx, hbKey, workerID, cfg.Worker.HeartbeatTTL)) which returns (bool,
error) and check the bool to detect existing heartbeat, or if you must use
SetArgs, call Result() on the StatusCmd and handle redis.Nil as the "already
exists" case, treat any non-nil error as a failure, and ensure a non-"OK" result
is also handled as an error.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In BUGS.md around lines 51 to 53, update the note about worker registry to

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360060998

- [review_comment] 2025-09-18T16:01:14Z by coderabbitai[bot] (BUGS.md:53)

```text
In BUGS.md around lines 51 to 53, update the note about worker registry to
document registry hygiene: specify that on startup or during periodic
maintenance the system must check each registered worker's heartbeat key and, if
the heartbeat key is missing or expired, remove that worker ID from the registry
(SREM) to prevent unbounded growth; avoid global SCAN/KEYS by iterating only the
registry set membership and checking each worker's heartbeat key, and clearly
state this pruning step and its frequency so implementers know to remove dead
workers when heartbeats are absent.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Replaced the non-atomic pipeline advice with a ZPOPMIN/Lua-based atomic mover. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Replaced the non-atomic pipeline advice with a ZPOPMIN/Lua-based atomic mover. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_009.md:653
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In BUGS.md around lines 81-115, the current mover pops entries with ZPOPMIN then

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061011

- [review_comment] 2025-09-18T16:01:14Z by coderabbitai[bot] (BUGS.md:115)

```text
In BUGS.md around lines 81-115, the current mover pops entries with ZPOPMIN then
uses a pipeline to re-add future items which can lose jobs if the pipeline
fails; replace the whole pop+pipe loop with a single server-side Lua script
executed via rdb.Eval that atomically moves due members (score <= now) from the
ZSET to the LIST up to a limit, passing schedKey and queueKey as KEYS and now
and limit as ARGV, then check the returned moved count and error and remove the
old loop and pipeline logic.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In cmd/admin-api/main.go around lines 32 to 35, the code checks the error return

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061022

- [review_comment] 2025-09-18T16:01:14Z by coderabbitai[bot] (cmd/admin-api/main.go:35)

```text
In cmd/admin-api/main.go around lines 32 to 35, the code checks the error return
from fs.Parse even though the FlagSet was created with flag.ExitOnError so Parse
will never return — remove the dead if-block and simply call
fs.Parse(os.Args[1:]) (or assign its result to _ if you prefer) without handling
an error; ensure no other logic depends on that removed branch and keep the
program flow unchanged.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In cmd/admin-api/main.go around line 59, the deferred call defer logger.Sync()

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061036

- [review_comment] 2025-09-18T16:01:15Z by coderabbitai[bot] (cmd/admin-api/main.go:59)

```text
In cmd/admin-api/main.go around line 59, the deferred call defer logger.Sync()
ignores its returned error; replace it with a deferred closure that captures and
checks the error (e.g. defer func(){ if err := logger.Sync(); err != nil {
fmt.Fprintf(os.Stderr, "logger sync error: %v\n", err) } }()), so Sync errors
are not silently swallowed and are written to stderr for visibility.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In cmd/admin-api/main.go around line 67 (and similarly for lines 98-112), the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061044

- [review_comment] 2025-09-18T16:01:15Z by coderabbitai[bot] (cmd/admin-api/main.go:67)

```text
In cmd/admin-api/main.go around line 67 (and similarly for lines 98-112), the
signal handler uses a hard-coded 5s timeout and never calls signal.Stop, which
can undercut cfg.ShutdownTimeout and leave the signal channel wired; change the
handler to use cfg.ShutdownTimeout (or derive timeout from ctx/cancel) when
creating the shutdown context, call signal.Notify on a channel and ensure you
call signal.Stop(ch) (preferably via defer) once the handler exits, and replace
the fixed time.After(5*time.Second) logic with a context-with-timeout using
cfg.ShutdownTimeout so the server shutdown honors configured timeout and the
notifier is cleaned up.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In cmd/admin-api/main.go around lines 69 to 71, remove the call to logger.Fatal

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061071

- [review_comment] 2025-09-18T16:01:15Z by coderabbitai[bot] (cmd/admin-api/main.go:71)

```text
In cmd/admin-api/main.go around lines 69 to 71, remove the call to logger.Fatal
which calls os.Exit and skips defers; instead log the error (e.g.
logger.Error/with context) and propagate a non-zero exit path so deferred
cleanup runs—either return the error from main and let os.Exit be called after
deferred cleanup or set an exitCode variable and call os.Exit(exitCode) only
after all defers have run; ensure Redis Close and logger.Sync defer calls remain
untouched.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In cmd/admin-api/main.go around lines 84 to 89, the current error check uses

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061081

- [review_comment] 2025-09-18T16:01:15Z by coderabbitai[bot] (cmd/admin-api/main.go:89)

```text
In cmd/admin-api/main.go around lines 84 to 89, the current error check uses
os.IsNotExist which does not detect Viper's missing-config error; replace that
check to detect viper.ConfigFileNotFoundError instead (e.g. use a type
assertion: if _, ok := err.(viper.ConfigFileNotFoundError); ok { return cfg, nil
} ) and return other errors as before; add the viper import if not already
present (and remove or keep os import only if still used elsewhere).
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In create_review_tasks.py around lines 100-113 (and also update the duplicate

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061102

- [review_comment] 2025-09-18T16:01:15Z by coderabbitai[bot] (create_review_tasks.py:113)

```text
In create_review_tasks.py around lines 100-113 (and also update the duplicate
entries around lines 138-143), the coverage threshold is inconsistent between
90% in the DoD and 80% in the task instructions; standardize both to 90%. Update
the task definition entries so any mention of coverage uses "90%+" (or a numeric
90) and remove or replace the 80% references to ensure both the criteria and
instructions are aligned to the 90% threshold.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In demos/responsive-tui.tape around lines 19-27, the script sets COLUMNS

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061123

- [review_comment] 2025-09-18T16:01:15Z by coderabbitai[bot] (demos/responsive-tui.tape:27)

```text
In demos/responsive-tui.tape around lines 19-27, the script sets COLUMNS
cosmetically (export COLUMNS=35) which doesn't actually resize the terminal;
change that to drive a real terminal width before the section by using a
terminal resize command (for example: stty cols 35 or resize -s <rows> 35) so
the TUI hits its breakpoint; make the same replacement in the other demo blocks
noted (lines 75-83, 133-141, 219-227) so each section begins with an actual
terminal width change rather than only exporting COLUMNS.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 7 | Demo now calls out COLUMNS overrides as cosmetic and trims long sleeps. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Demo now calls out COLUMNS overrides as cosmetic and trims long sleeps. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_001.md:750
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In demos/responsive-tui.tape around lines 73-74 (also apply same fix at 131-132,

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061143

- [review_comment] 2025-09-18T16:01:15Z by coderabbitai[bot] (demos/responsive-tui.tape:74)

```text
In demos/responsive-tui.tape around lines 73-74 (also apply same fix at 131-132,
217-218, 311-313): the test sets a fake COLUMNS environment variable but does
not restore the original value, leaking the fake into downstream steps; modify
each section to save the original value (e.g., prev="$COLUMNS" or detect unset),
set the fake COLUMNS for the test, and then after the section restore it by
exporting COLUMNS="$prev" if prev was set or by unsetting COLUMNS if prev was
originally unset so downstream steps see the original environment.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In dependency_analysis.py around lines 302 to 324, the "provides" entries are

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061166

- [review_comment] 2025-09-18T16:01:16Z by coderabbitai[bot] (dependency_analysis.py:324)

```text
In dependency_analysis.py around lines 302 to 324, the "provides" entries are
currently copied verbatim while hard/soft/enables are normalized and aliased,
which can cause identifier mismatches; update the function to normalize and
resolve aliases for each item in the "provides" list the same way as the other
dependency lists (e.g., map each dep through normalize_name and resolve_alias)
or, if you intentionally want them only for display, add an explicit
comment/docstring near the function noting that "provides" is display-only and
must not be used for dependency resolution; choose one approach and apply it
consistently.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/scripts/test-staging-deployment.sh around lines 441 to 449, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061179

- [review_comment] 2025-09-18T16:01:16Z by coderabbitai[bot] (deployments/scripts/test-staging-deployment.sh:449)

```text
In deployments/scripts/test-staging-deployment.sh around lines 441 to 449, the
monitoring namespace is hardcoded to "monitoring"; change this to a parameter by
adding a -m/--monitoring-ns option to the script's argument parsing (matching
the style used in setup-monitoring.sh), introduce a local variable (e.g.,
monitoring_ns) set from that option with a default of "monitoring", update the
help/usage text, and ensure all checks and add_test_result calls use that
variable so existing behavior stays the same when the flag is not provided.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## docs/07_test_plan.md lines 28–37: the chaos tests as written require

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061205

- [review_comment] 2025-09-18T16:01:16Z by coderabbitai[bot] (docs/07_test_plan.md:37)

```text
docs/07_test_plan.md lines 28–37: the chaos tests as written require
CAP_NET_ADMIN and will fail on GitHub-hosted runners; update the document to (1)
note the CAP_NET_ADMIN requirement and explicitly gate/skip these steps on
hosted runners, (2) provide alternatives and examples: run netem in a privileged
sidecar container with CAP_NET_ADMIN, run the tests on self-hosted runners that
grant the capability, or replace host-level injections with proxy-based tools
(toxiproxy/pumba) that work on hosted runners, and (3) add cleanup and detection
guidance so CI can detect capability absence and automatically skip these steps
while pointing users to the privileged-run instructions.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/07_test_plan.md around lines 53 to 57, the CPU governor step assumes

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061220

- [review_comment] 2025-09-18T16:01:16Z by coderabbitai[bot] (docs/07_test_plan.md:57)

```text
In docs/07_test_plan.md around lines 53 to 57, the CPU governor step assumes
cpupower is present and runnable with sudo; on stock GitHub runners cpupower may
not exist and the step will fail — change the instructions to make this
best‑effort by first checking for cpupower (command -v cpupower) and only
attempting sudo cpupower frequency-set -g performance when available, allowing
the command to fail silently (e.g., || true), and apply the same guarded
approach for restoring the governor on exit so the job won’t fail if cpupower is
absent or non‑runnable.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/15_promotion_checklists.md around line 3, the "Last updated" timestamp

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061235

- [review_comment] 2025-09-18T16:01:16Z by coderabbitai[bot] (docs/15_promotion_checklists.md:3)

```text
In docs/15_promotion_checklists.md around line 3, the "Last updated" timestamp
is stale (2025-09-12); update that line to the current date (2025-09-18) so the
document reflects the correct last-updated date.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/calendar-view.md around lines 63 to 77, the RecurringRule sample

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061257

- [review_comment] 2025-09-18T16:01:17Z by coderabbitai[bot] (docs/api/calendar-view.md:77)

```text
In docs/api/calendar-view.md around lines 63 to 77, the RecurringRule sample
mixes Go's time.Duration with JSON string values (e.g. "300s"); update the
documentation and sample struct so JSON shows a string for Jitter (or introduce
a custom Duration type) — either change the Jitter field to string in the sample
JSON struct and use a string example like "300s", or, if the SDK should keep
strong typing, define a custom Duration type that marshals/unmarshals to/from a
JSON string and replace time.Duration with that type in the sample; ensure the
docs show the final chosen representation and include an example value formatted
as a string.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/calendar-view.md around lines 198 to 206, the request example

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061274

- [review_comment] 2025-09-18T16:01:17Z by coderabbitai[bot] (docs/api/calendar-view.md:206)

```text
In docs/api/calendar-view.md around lines 198 to 206, the request example
includes a client-supplied identity (user_id) and header examples reference
X-User-ID; remove the user_id field from the JSON example and remove any
X-User-ID header examples, and update any explanatory text to state that
identity is derived from the bearer token (Authorization: Bearer ...) instead;
apply the same changes to the other affected ranges (lines 225-241 and 666-681).
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/canary-deployments.md around lines 558 to 597, replace all example

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061288

- [review_comment] 2025-09-18T16:01:17Z by coderabbitai[bot] (docs/api/canary-deployments.md:597)

```text
In docs/api/canary-deployments.md around lines 558 to 597, replace all example
deployment IDs that use legacy placeholders like "canary_..." with valid
26-character ULIDs (Crockford base32, uppercase) that match the regex
^[0-9A-HJKMNP-TV-Z]{26}$; update the specific occurrences mentioned at lines 70,
369, 376, and 783 as well as any other examples in the document, ensuring each
ID is 26 uppercase base32 characters and adjust any accompanying validation text
or examples to use those ULIDs.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/chaos-harness.md around lines 430 to 441, document the semantics of

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061302

- [review_comment] 2025-09-18T16:01:17Z by coderabbitai[bot] (docs/api/chaos-harness.md:441)

```text
In docs/api/chaos-harness.md around lines 430 to 441, document the semantics of
the wildcard used in examples: explicitly state that scope_value: "*" means “all
workers” (and similarly for queue scope), clarify its precedence relative to
specific IDs/names (e.g., exact matches take precedence over the wildcard),
state whether other wildcard patterns or full regex are supported or not (and if
supported, give syntax and matching rules), and add a short example and a note
right after the “Worker Scope” section showing usage and precedence so readers
aren’t left to guess.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/chaos-harness.md around lines 601 to 626, the document lacks a

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061316

- [review_comment] 2025-09-18T16:01:17Z by coderabbitai[bot] (docs/api/chaos-harness.md:626)

```text
In docs/api/chaos-harness.md around lines 601 to 626, the document lacks a
clear, global timestamp format; update this section to state that all timestamps
use RFC 3339 in UTC (e.g., 2025-09-18T14:30:00Z) and add a single sentence near
the top of the Metrics/Report/Troubleshooting areas noting "Timestamps are RFC
3339 in UTC" so examples and time-series graphs uniformly follow that format.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/PRD.md around lines 119 to 124, the heartbeat key contains a stray

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061339

- [review_comment] 2025-09-18T16:01:18Z by coderabbitai[bot] (docs/PRD.md:124)

```text
In docs/PRD.md around lines 119 to 124, the heartbeat key contains a stray
backtick/space and the key formatting is inconsistent; remove the extra
backtick/space from `jobqueue:processing:worker:<ID> ` and make all keys
consistently formatted (e.g., wrap each key in backticks without trailing
spaces) so the list entries are uniform.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/SLAPS/worker-reflections/claude-001-reflection.md around lines 111 to

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061357

- [review_comment] 2025-09-18T16:01:18Z by coderabbitai[bot] (docs/SLAPS/worker-reflections/claude-001-reflection.md:114)

```text
In docs/SLAPS/worker-reflections/claude-001-reflection.md around lines 111 to
114, the footer contains metadata as prose which is not machine-readable; move
each metadata item (end of reflection marker, total time in SLAPS, tasks
completed, primary lesson learned) into the document front-matter as explicit
fields (e.g., end_reflection, total_time, tasks_completed, primary_lesson) and
remove the corresponding footer lines at 111–114 so tooling can index the
values.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/SLAPS/worker-reflections/claude-008-reflection.md around lines 1 to 4,

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061395

- [review_comment] 2025-09-18T16:01:18Z by coderabbitai[bot] (docs/SLAPS/worker-reflections/claude-008-reflection.md:4)

```text
In docs/SLAPS/worker-reflections/claude-008-reflection.md around lines 1 to 4,
the front matter only contains date and worker_id; add explicit front-matter
fields required by the docs engine: title, slug, and description, and include a
mermaid: true/false toggle (or mermaid: enabled/disabled) as needed. Populate
title with a concise human-readable title, slug with a URL-safe identifier,
description with a one-line summary, and set the mermaid flag to the appropriate
boolean so the docs generator can parse and render diagrams.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## docs/SLAPS/worker-reflections/claude-008-reflection.md around lines 27-33:

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061415

- [review_comment] 2025-09-18T16:01:18Z by coderabbitai[bot] (docs/SLAPS/worker-reflections/claude-008-reflection.md:33)

```text
docs/SLAPS/worker-reflections/claude-008-reflection.md around lines 27-33:
remove the vanity line-count mentions and replace them with direct links to the
real artifacts (architecture docs, OpenAPI 3.0 spec, JSON Schema definitions)
using repo-relative or absolute URLs; for each linked artifact add a one-line
description (what it is and where to find key sections, e.g., "architecture:
design/architecture/job-genealogy.md — sections on algorithms and UX") and
ensure links point to the canonical files or rendered docs, not just folders.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## docs/SLAPS/worker-reflections/claude-008-reflection.md lines 37-41: the strategy

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061431

- [review_comment] 2025-09-18T16:01:18Z by coderabbitai[bot] (docs/SLAPS/worker-reflections/claude-008-reflection.md:41)

```text
docs/SLAPS/worker-reflections/claude-008-reflection.md lines 37-41: the strategy
bullets are generic and need concrete cross-references; update each bullet to
include a relative link or path to an example where you applied that strategy
(e.g., replace "Read Everything First" with "Read Everything First — review
notes: docs/reviews/P4.T044-notes.md", "Comprehensive Test Planning" with
"Comprehensive Test Planning — test plan: tests/plans/P4.T044.md", "Mermaid
Diagrams Everywhere" with "Mermaid Diagrams — diagram:
docs/diagrams/genealogy.mmd", and "Four-Part Adaptation Plan" with a pointer to
the workflow doc or task where you executed it), ensuring links are valid
relative paths and keep the bullet wording concise.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/SLAPS/worker-reflections/claude-008-reflection.md around lines 48–51,

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061453

- [review_comment] 2025-09-18T16:01:19Z by coderabbitai[bot] (docs/SLAPS/worker-reflections/claude-008-reflection.md:51)

```text
In docs/SLAPS/worker-reflections/claude-008-reflection.md around lines 48–51,
the recommendations are high-level and lack ownership and measurable follow-up;
add an "Action Items" section listing each recommendation as a checkboxed task
with an assigned owner, clear acceptance criteria, and a deadline (e.g., owner:
@username, acceptance: what success looks like, due: YYYY-MM-DD). Convert each
bullet into an actionable item (stabilize shared test infra; add lightweight
signals; define prerequisite chains), and add a final line offering to draft the
issue and PR templates (or link to them) so someone can immediately pick this
up. Ensure formatting is consistent with the doc (checkboxes, owner, acceptance
criteria, deadline) so these items are testable and trackable.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/YOU ARE WORKER 6.md around lines 8-9 (and also lines 27-29), the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061475

- [review_comment] 2025-09-18T16:01:19Z by coderabbitai[bot] (docs/YOU ARE WORKER 6.md:9)

```text
In docs/YOU ARE WORKER 6.md around lines 8-9 (and also lines 27-29), the
instructions mistakenly reference claude-001; update the path examples to
claude-006 for this worker (replace any "claude-001" occurrences with
"claude-006"), and ensure any related guidance about choosing 001/002/003
reflects the correct worker-specific naming (use 006 for this doc).
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 20 to 26, the documented test

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061522

- [review_comment] 2025-09-18T16:01:19Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:26)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 20 to 26, the documented test
names include a non-existent TestWebhookDeliveryWithRetries; update the block to
list the actual test names: keep TestHMACSigner_SignPayload and
TestHMACSigner_VerifySignature as-is, and replace the TestBackoffScheduler_*
entry with the concrete names TestBackoffScheduler_ExponentialStrategy,
TestBackoffScheduler_LinearStrategy, and TestBackoffScheduler_FixedStrategy;
remove or replace TestWebhookDeliveryWithRetries with
TestWebhookHarness_RetryOnFailure (test/integration/webhook_harness_test.go:405)
or use a generic TestWebhookHarness_* selector listing the real examples
(BasicDelivery, RetryOnFailure, ConcurrentDeliveries, SignatureValidation) so
the documentation matches existing test names and file locations.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 246 to 265, the perf tables

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061546

- [review_comment] 2025-09-18T16:01:19Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:265)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 246 to 265, the perf tables
lack traceability metadata; update each table row (or add a single preface block
above the Unit and Integration test tables) to include the git commit SHA used,
the exact random test seed (if any), and exact tool/runtime versions (Go
version, OS, CPU/host, Redis/NATS/docker image tags, TLS/other flags). For each
row either append columns or add parenthetical metadata that lists: commit:
<full SHA>, seed: <value or "n/a">, tooling: Go <x.y.z>, OS <name + version>,
CPU <model>, Redis <version+source>, NATS <version+source>, Docker <version if
used>, and a path to the persisted raw output file (e.g.,
benchmarks/event-hooks/latest.txt or artifacts/...). Ensure format is consistent
across all rows and include the exact string values (not ranges or approximate
names).
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 250–255 (and also update lines

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061566

- [review_comment] 2025-09-18T16:01:19Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:255)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 250–255 (and also update lines
262–264), the Go version in the performance table is stale (shows "Go 1.22.5")
while the repo/CI/Docker use Go 1.25; change those occurrences to "Go 1.25" or
replace with an explicit reference to the repository/CI toolchain (e.g., "Go
1.25 (as specified in go.mod/GitHub Actions/Dockerfiles)") so the docs match the
actual toolchain.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 252 to 256, the note claiming

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061582

- [review_comment] 2025-09-18T16:01:20Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:256)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 252 to 256, the note claiming
“Captured with `BENCH_MEM=1`” is incorrect and the reproduce instruction is
vague; update the table's note to state the real flag `-benchmem` (e.g.,
"Captured with `-benchmem` to record allocations"), and modify the reproduce
paragraph to show concrete, reproducible commands using `go test -bench=...
-benchtime=... -benchmem > benchmarks/event-hooks/latest.txt` (or one file per
benchmark), ensuring the examples include the exact `-bench` pattern,
`-benchtime`, `-benchmem`, and redirection to persist raw output alongside the
commit.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 331 to 355, the example test

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360061598

- [review_comment] 2025-09-18T16:01:20Z by coderabbitai[bot] (EVENT_HOOKS_TEST_DOCUMENTATION.md:355)

```text
In EVENT_HOOKS_TEST_DOCUMENTATION.md around lines 331 to 355, the example test
calls NewEventFilter() without showing required configuration or explaining
defaults; update the example to either (a) instantiate NewEventFilter with a
minimal, realistic config object used by the project (show required fields and
values inline) and use that instance in the test, or (b) if NewEventFilter
supports a no-arg/DefaultEventFilter, replace the call and comment that defaults
are used; ensure the example compiles and the t.Run block contains a concrete
assertion using that configured filter rather than a placeholder.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In .claude/agents/feature-enhancer.md lines 1-6, the file triggers markdownlint

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066642

- [review_comment] 2025-09-18T16:02:29Z by coderabbitai[bot] (.claude/agents/feature-enhancer.md:6)

```text
In .claude/agents/feature-enhancer.md lines 1-6, the file triggers markdownlint
rule MD013 (line length); either reflow the long description lines to reasonable
lengths (wrap paragraphs or break the long single-line description into multiple
shorter lines) or explicitly disable the rule for this file by adding <!--
markdownlint-disable MD013 --> at the top before the content and then add <!--
markdownlint-enable MD013 --> at EOF to re-enable the rule; ensure no other
lines exceed the project's max line length if you choose wrapping.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In cmd/job-queue-system/main.go around lines 182 to 189, the command prints a

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066665

- [review_comment] 2025-09-18T16:02:29Z by coderabbitai[bot] (cmd/job-queue-system/main.go:189)

```text
In cmd/job-queue-system/main.go around lines 182 to 189, the command prints a
plain string ("dead letter queue purged") after successfully purging the DLQ;
change this to emit a machine-readable JSON success object instead (consistent
with other commands). Replace the fmt.Println call with JSON output to stdout
(e.g., an object with keys like "status":"ok" and "message":"dead letter queue
purged" or similar), using the standard library JSON encoder to write to
os.Stdout and return the same exit behavior; keep existing error handling
(logger.Fatal) unchanged.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/admin-api/deploy.sh around lines 101-107 you currently print a

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066685

- [review_comment] 2025-09-18T16:02:29Z by coderabbitai[bot] (deployments/admin-api/deploy.sh:107)

```text
In deployments/admin-api/deploy.sh around lines 101-107 you currently print a
readiness failure and continue; change the else branch so the script exits
non-zero (e.g., echo the failure to stderr and run exit 1) so a failed readiness
check fails the run; alternatively ensure the script is running with set -e and
propagate the curl failure, but the minimal fix is to add an exit 1 in the else
path after printing the failure.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/docker/docker-compose.yaml around lines 20 to 56 (and also apply

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066702

- [review_comment] 2025-09-18T16:02:29Z by coderabbitai[bot] (deployments/docker/docker-compose.yaml:56)

```text
In deployments/docker/docker-compose.yaml around lines 20 to 56 (and also apply
same change to lines 57 to 84), the service blocks lack restart policies so
containers won’t automatically recover; add a restart policy to each application
service (for example restart: unless-stopped or restart: always) directly under
the service definition (align with other top-level keys such as ports/env_file)
and, if needed, add restart_policy options for finer control
(maximum_retry_count, window) to ensure services are automatically restarted on
failure or Docker daemon restarts.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/docker/rbac-configs/roles.yaml around lines 107 to 116, the RBAC

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066723

- [review_comment] 2025-09-18T16:02:30Z by coderabbitai[bot] (deployments/docker/rbac-configs/roles.yaml:116)

```text
In deployments/docker/rbac-configs/roles.yaml around lines 107 to 116, the RBAC
role assignment comments are vague about who enforces rules, when they are
applied, and the precedence; add a clear documentation block named
role_assignment_rules immediately above or beside the existing domain rules that
states: Enforced by: RBAC Token Service during token issuance; Precedence:
explicit user assignment > group membership > domain rules > default_role;
Evaluation: token email domain matched against rules in order. Ensure the block
is concise, YAML-comment friendly (or a YAML mapping if you prefer
machine-readable docs), and placed where maintainers reading the file will see
it before the domain entries.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/kubernetes/rbac-monitoring.yaml around lines 33 to 43, the alert

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066741

- [review_comment] 2025-09-18T16:02:30Z by coderabbitai[bot] (deployments/kubernetes/rbac-monitoring.yaml:43)

```text
In deployments/kubernetes/rbac-monitoring.yaml around lines 33 to 43, the alert
expression using up{app="rbac-token-service"} == 0 is brittle and can miss cases
where all targets are absent; replace the expr with an aggregate/absent-aware
expression such as using sum() and absent(), e.g. trigger when
absent(up{app="rbac-token-service"}) or when sum(up{app="rbac-token-service"})
== 0 so that total outages and aggregated down states are detected; keep the
same for/labels/annotations but update the expr to the combined absent OR sum
condition.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Runbook link now points to the documented ops guide. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Runbook link now points to the documented ops guide. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:380
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/kubernetes/rbac-monitoring.yaml around lines 45 to 59, the alert

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066756

- [review_comment] 2025-09-18T16:02:30Z by coderabbitai[bot] (deployments/kubernetes/rbac-monitoring.yaml:59)

```text
In deployments/kubernetes/rbac-monitoring.yaml around lines 45 to 59, the alert
expression mixes different label keys (job vs app) which can cause silent
mismatches; update the rule to consistently use job="rbac-token-service"
everywhere in the expression (both the error rate numerator and the total
request denominator), and then search and update dashboard panels/targets to use
the same job="rbac-token-service" selector (or switch all to a chosen canonical
selector such as service=<name> across alerts, recording rules and dashboards)
so all queries use the identical label key/value.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/kubernetes/rbac-token-service-deployment.yaml around lines

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066805

- [review_comment] 2025-09-18T16:02:31Z by coderabbitai[bot] (deployments/kubernetes/rbac-token-service-deployment.yaml:151)

```text
In deployments/kubernetes/rbac-token-service-deployment.yaml around lines
146–151 the CORS allowed_origins are hardcoded to staging/prod domains; replace
this with a parameterized source by loading origins from an environment variable
or ConfigMap (e.g., ORIGINS_CSV) or via your Helm/template values, have the app
parse the CSV into the allowed_origins array at startup, provide a sensible
default/fallback and document how to set the env/config, and ensure
allowed_methods and other CORS fields remain populated from the same
configurable source.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/kubernetes/rbac-token-service-deployment.yaml lines 199-201 (and

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066815

- [review_comment] 2025-09-18T16:02:31Z by coderabbitai[bot] (deployments/kubernetes/rbac-token-service-deployment.yaml:201)

```text
In deployments/kubernetes/rbac-token-service-deployment.yaml lines 199-201 (and
also 263-267), the pod securityContext uses UID/GID 1000 which may collide with
host users. Update runAsUser and fsGroup to a high, non-host ID like 10001
(consistent across all containers and pod-level securityContext). Keep
runAsNonRoot: true. Verify any related runAsGroup or container-level overrides
also use 10001, and ensure both affected blocks are updated.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/scripts/deploy-rbac-staging.sh around line 11, the dynamic source

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066829

- [review_comment] 2025-09-18T16:02:31Z by coderabbitai[bot] (deployments/scripts/deploy-rbac-staging.sh:11)

```text
In deployments/scripts/deploy-rbac-staging.sh around line 11, the dynamic source
of "${SCRIPT_DIR}/lib/logging.sh" triggers ShellCheck SC1091; add a ShellCheck
source hint comment immediately above the source line to point to the actual
path of the file in the repo (for example: # shellcheck
source=deployments/scripts/lib/logging.sh) so ShellCheck can resolve it, then
keep the existing source "${SCRIPT_DIR}/lib/logging.sh" unchanged.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/scripts/deploy-rbac-staging.sh around lines 40-47 (and also apply

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066843

- [review_comment] 2025-09-18T16:02:31Z by coderabbitai[bot] (deployments/scripts/deploy-rbac-staging.sh:47)

```text
In deployments/scripts/deploy-rbac-staging.sh around lines 40-47 (and also apply
same fix to 116-123 and 126-132), you build a Docker image locally but never
push or load it so a remote cluster will get ImagePullBackOff; either push the
built image to a registry and ensure rbac-token-service-deployment.yaml
references the exact same image name/tag (use $IMAGE_NAME consistently and push
tags like :staging), or if targeting a local cluster (kind/minikube) replace the
push step with loading the image into the cluster via kind load docker-image
"$IMAGE_NAME" or minikube image load "$IMAGE_NAME"; pick one approach and make
the script perform the corresponding push or load and keep the deployment
YAML/image variable consistent.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/scripts/deploy-staging.sh around lines 52 to 64, the file defines

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066855

- [review_comment] 2025-09-18T16:02:31Z by coderabbitai[bot] (deployments/scripts/deploy-staging.sh:64)

```text
In deployments/scripts/deploy-staging.sh around lines 52 to 64, the file defines
duplicate log_info/log_warn/log_error helpers; remove these local definitions
and instead source the shared logging lib (deployments/scripts/lib/logging.sh)
before any logging is used. Add a check that the logging.sh file exists and
source it (or exit with an error if missing) so the script fails fast when the
shared helper is unavailable; do not reimplement the functions locally to avoid
drift.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/scripts/deploy-staging.sh around lines 100 to 118, the IMAGE_NAME

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066877

- [review_comment] 2025-09-18T16:02:31Z by coderabbitai[bot] (deployments/scripts/deploy-staging.sh:118)

```text
In deployments/scripts/deploy-staging.sh around lines 100 to 118, the IMAGE_NAME
is built without a required registry namespace which can yield
docker.io/<app>:tag (the implicit "library" namespace) and cause push failures;
validate that DOCKER_NAMESPACE is set (fail fast if missing), construct a
fully-qualified IMAGE_NAME combining REGISTRY (trim trailing slash) +
DOCKER_NAMESPACE + APP_NAME + IMAGE_TAG, and use that canonical IMAGE_NAME
everywhere (build, push, and later kubectl set image) rather than embedding
"$REGISTRY/$APP_NAME:$IMAGE_TAG" inline; also audit
deployments/admin-api/deploy.sh and CI workflows (.github/workflows/*) to
standardize on REGISTRY + DOCKER_NAMESPACE + APP_NAME and add the same
validation where applicable.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/scripts/deploy-staging.sh around lines 147-155, the script uses a

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066887

- [review_comment] 2025-09-18T16:02:32Z by coderabbitai[bot] (deployments/scripts/deploy-staging.sh:155)

```text
In deployments/scripts/deploy-staging.sh around lines 147-155, the script uses a
direct if [ $? -ne 0 ] conditional to check rollout status; replace this with an
explicit capture of the command's exit code immediately after running the
rollout/status command (e.g., run kubectl rollout status ... --timeout=... and
store its exit code in a variable), then test that variable (if [ "$status" -ne
0 ]) to decide logging, calling rollback, and exiting; ensure the error log
includes the captured exit code or command output for clarity and use that same
code in exit so the caller sees the actual failure code.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/scripts/deploy-staging.sh around lines 182 to 186 there is a

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066896

- [review_comment] 2025-09-18T16:02:32Z by coderabbitai[bot] (deployments/scripts/deploy-staging.sh:186)

```text
In deployments/scripts/deploy-staging.sh around lines 182 to 186 there is a
duplicate call to register_port_forward "$PF_PID"; remove the repeated line so
the port-forward PID is only registered once. Keep the single
register_port_forward "$PF_PID" immediately after PF_PID=$! (and before sleep 5)
to ensure the background PID is recorded exactly once; do not change the
surrounding port-forward or PID assignment logic.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/scripts/health-check-rbac.sh around lines 41 to 44, the kubectl

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066908

- [review_comment] 2025-09-18T16:02:32Z by coderabbitai[bot] (deployments/scripts/health-check-rbac.sh:44)

```text
In deployments/scripts/health-check-rbac.sh around lines 41 to 44, the kubectl
cluster-info call can hang CI; wrap it with a timeout (e.g. 10s) and fail if it
exceeds that. Implement: if command -v timeout >/dev/null use timeout 10s
kubectl cluster-info, else use kubectl cluster-info --request-timeout=10s; keep
the existing error message and return 1 when the guarded call fails or times
out.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/scripts/health-check-rbac.sh around lines 354 to 380, the parsed

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066924

- [review_comment] 2025-09-18T16:02:32Z by coderabbitai[bot] (deployments/scripts/health-check-rbac.sh:380)

```text
In deployments/scripts/health-check-rbac.sh around lines 354 to 380, the parsed
TIMEOUT value isn't validated and non-integer input will cause arithmetic
failures later; after assigning TIMEOUT in parse_args (or immediately after
parse_args returns) validate it with a simple integer check (e.g. regex like
^[0-9]+$) and ensure it's positive, and if the check fails print a clear error
message including the invalid value and exit 1 so the script fails fast on
garbage input.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/scripts/setup-monitoring.sh around line 11, ShellCheck is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066944

- [review_comment] 2025-09-18T16:02:32Z by coderabbitai[bot] (deployments/scripts/setup-monitoring.sh:11)

```text
In deployments/scripts/setup-monitoring.sh around line 11, ShellCheck is
complaining about sourcing a file (SC1091); add an explicit source hint comment
immediately above the source line to satisfy CI. Place a ShellCheck source
directive that points to the referenced file (for example: # shellcheck
source=./lib/logging.sh) on the line above the existing source
"${SCRIPT_DIR}/lib/logging.sh" so the linter recognizes the target and the
warning is silenced.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/scripts/setup-monitoring.sh around lines 30-31 (and the similar

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066956

- [review_comment] 2025-09-18T16:02:32Z by coderabbitai[bot] (deployments/scripts/setup-monitoring.sh:31)

```text
In deployments/scripts/setup-monitoring.sh around lines 30-31 (and the similar
block at 194-209), the ALERTMANAGER_WEBHOOK_URL is defaulting to
http://localhost:9093/webhook which is unsafe/useless; change the logic to fail
fast when ALERTMANAGER_WEBHOOK_URL is unset or validate it and reject
localhost/loopback addresses. Specifically, remove the localhost default, check
if ALERTMANAGER_WEBHOOK_URL is non-empty, verify it looks like a sane URL
(http(s) scheme and host not localhost/127.0.0.1/::1), and exit with a clear
error if the check fails so the script requires a real webhook URL instead of
silently using localhost.
```

> [!INFO]- **Accepted**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | Yes | No | 8 | Monitoring script sources the shared logging helpers instead of duplicating them. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> Monitoring script sources the shared logging helpers instead of duplicating them. Evidence: docs/audits/code-reviews/PR3/e35da518e543d331abf0b57fa939d682d39f5a88.md.chunk_006.md:896
>
> **Alternatives Considered**
> Not discussed.
>
> **Lesson(s) Learned**
> None recorded.

## In deployments/scripts/setup-monitoring.sh around lines 117 to 120, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066974

- [review_comment] 2025-09-18T16:02:33Z by coderabbitai[bot] (deployments/scripts/setup-monitoring.sh:120)

```text
In deployments/scripts/setup-monitoring.sh around lines 117 to 120, the
SCRIPT_DIR variable is being computed twice (once globally and again inside a
function); remove the duplicate computation by keeping the existing top-level
SCRIPT_DIR assignment and deleting the redundant assignment within the function,
then ensure any code inside the function references the top-level SCRIPT_DIR
variable (no redefinition) so the script still resolves paths correctly.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/scripts/setup-monitoring.sh around lines 146–214, the script

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360066980

- [review_comment] 2025-09-18T16:02:33Z by coderabbitai[bot] (deployments/scripts/setup-monitoring.sh:214)

```text
In deployments/scripts/setup-monitoring.sh around lines 146–214, the script
checks for a secret named "alertmanager-main" but creates
"alertmanager-rbac-config", so Prometheus Operator will ignore the config;
change the creation/patch step to create or update the secret name the operator
expects (e.g., create/patch "alertmanager-main" in $MONITORING_NAMESPACE with
the generated alertmanager config) and ensure the secret key matches the
operator’s expected key (replace "alertmanager-rbac-config" with
"alertmanager-main" or vice‑versa consistently, using kubectl create secret
generic alertmanager-main --from-literal=alertmanager.yml="$alertmanager_config"
--dry-run=client -o yaml | kubectl apply -f - or perform a kubectl patch if you
need to update an existing secret).
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/scripts/test-staging-deployment.sh around lines 278 to 280,

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360067025

- [review_comment] 2025-09-18T16:02:33Z by coderabbitai[bot] (deployments/scripts/test-staging-deployment.sh:280)

```text
In deployments/scripts/test-staging-deployment.sh around lines 278 to 280,
remove the blind "sleep 5" and replace it with a socket polling loop that
repeatedly checks the forwarded local port until it accepts connections or a
configurable timeout is reached; implement the loop using a lightweight check
(e.g. nc -z, bash /dev/tcp/host/port, or timeout+curl) with short sleeps between
attempts, fail the script with a clear error if the port never becomes available
within the timeout, and only proceed to the HTTP checks once the socket is
confirmed open.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/scripts/test-staging-deployment.sh around lines 281 to 314, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360067037

- [review_comment] 2025-09-18T16:02:33Z by coderabbitai[bot] (deployments/scripts/test-staging-deployment.sh:314)

```text
In deployments/scripts/test-staging-deployment.sh around lines 281 to 314, the
timeout command uses a hard-coded 30 seconds in two places; replace those
literals with the configured TIMEOUT variable (e.g. timeout "$TIMEOUT" ...) so
the script honors the configured timeout consistently for both health and
metrics checks, ensuring proper quoting of the variable and keeping the same
bash -c loop structure.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/scripts/test-staging-deployment.sh around lines 321 to 376, the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360067046

- [review_comment] 2025-09-18T16:02:33Z by coderabbitai[bot] (deployments/scripts/test-staging-deployment.sh:376)

```text
In deployments/scripts/test-staging-deployment.sh around lines 321 to 376, the
RBAC tests start kubectl port-forward and then sleep 5 which can race and cause
test flakiness; replace the static sleep with a readiness poll that waits for
the local service to respond (e.g., loop up to a timeout calling curl -sS
--max-time 1 http://localhost:8081/health or the auth validate endpoint and
break when it returns 200/expected body), retrying every 1s and failing after a
configurable timeout; keep the port-forward in background, preserve
cleanup_port_forward/trap, and after the poll proceed with the bootstrap-token
retrieval and token-based tests (mark failure if readiness timeout occurs before
running tests).
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/12_performance_baseline.md around line 3, the "Last updated" stamp is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360067063

- [review_comment] 2025-09-18T16:02:34Z by coderabbitai[bot] (docs/12_performance_baseline.md:3)

```text
In docs/12_performance_baseline.md around line 3, the "Last updated" stamp is
stale; update the date to reflect this commit by replacing "Last updated:
2025-09-12" with the current commit date (or the intended release/merge date) so
the file shows the correct last-updated timestamp.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/12_performance_baseline.md around lines 51-53, the doc runs

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360067078

- [review_comment] 2025-09-18T16:02:34Z by coderabbitai[bot] (docs/12_performance_baseline.md:53)

```text
In docs/12_performance_baseline.md around lines 51-53, the doc runs
./bin/job-queue-system without explaining how that binary is produced; add a
prerequisite build step immediately before "3) In one shell, run the worker"
that instructs readers to build the binary (for example: run make build or run
go build ./cmd/job-queue-system -o ./bin/job-queue-system) and mention the
resulting path ./bin/job-queue-system so users don’t have to guess.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/12_performance_baseline.md around lines 64-65, replace the vague "curl

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360067080

- [review_comment] 2025-09-18T16:02:34Z by coderabbitai[bot] (docs/12_performance_baseline.md:65)

```text
In docs/12_performance_baseline.md around lines 64-65, replace the vague "curl
/metrics" note with a concrete one-liner that sets METRICS_URL to the binary's
default metrics address and shows a copy-pasteable curl that saves metrics to a
timestamped file; to do this, inspect the binary's flag parsing to determine the
actual default for --metrics-addr and use that host:port in the METRICS_URL
default (replace 9091 if the code's default is different), then add the two
lines: METRICS_URL=${METRICS_URL:-http://<actual-default>/metrics}  # set to
your --metrics-addr and curl -fsSL "$METRICS_URL" | tee "metrics_$(date
+%s).prom".
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/admin-api.md around lines 360 to 366, the health endpoint is

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360067091

- [review_comment] 2025-09-18T16:02:34Z by coderabbitai[bot] (docs/api/admin-api.md:366)

```text
In docs/api/admin-api.md around lines 360 to 366, the health endpoint is
documented as /health but the codebase and deployment use /healthz; update the
documentation to use /healthz (and similarly mention /readyz where appropriate)
so probes and runbooks match: replace occurrences of /health with /healthz in
this section and verify the example HTTP request and response remain unchanged.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/capacity-planning-api.md around lines 322-324, the example import

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360067105

- [review_comment] 2025-09-18T16:02:34Z by coderabbitai[bot] (docs/api/capacity-planning-api.md:324)

```text
In docs/api/capacity-planning-api.md around lines 322-324, the example import
references an internal package
("github.com/flyingrobots/go-redis-work-queue/internal/automatic-capacity-planning")
which cannot be imported outside its module; either change the import to a
public package path (move the package out of internal or point to a published
public module) or add a clear note immediately above the snippet that this
example must be placed inside that repository/module tree (so readers know it
won’t work from external modules). Ensure the doc shows the correct public
import path or the placement caveat, and remove the misleading “Replace the
import above…” line if you opt for the placement note.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/FEATURE_ENHANCEMENT_AGENT_PROMPT.md around lines 156 to 161 the link

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360067121

- [review_comment] 2025-09-18T16:02:34Z by coderabbitai[bot] (docs/FEATURE_ENHANCEMENT_AGENT_PROMPT.md:161)

```text
In docs/FEATURE_ENHANCEMENT_AGENT_PROMPT.md around lines 156 to 161 the link
label reads docs/colors.yml while the actual markdown link target points to
colors.yml, causing a mismatch; update the link target to point to
docs/colors.yml (or alternatively change the label to colors.yml) so label and
target match and verify the relative path is correct from this file's location.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/PRD.md around lines 154–156, the phrase "scan" implies using Redis KEYS

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360067133

- [review_comment] 2025-09-18T16:02:35Z by coderabbitai[bot] (docs/PRD.md:156)

```text
In docs/PRD.md around lines 154–156, the phrase "scan" implies using Redis KEYS
which can block; update the text to mandate using Redis SCAN with a MATCH
pattern and a COUNT parameter and describe a cursor-based loop with safe bounds
(e.g., iteration limits) and incremental backoff between iterations to avoid
overwhelming Redis; specify using MATCH for heartbeat key pattern, set a
reasonable COUNT value, resume from the returned cursor until zero, and include
guidance to back off (sleep) when SCAN returns many keys or after a full pass to
prevent blocking and thrashing.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In Makefile around line 6: the LDFLAGS uses -X main.version=$(VERSION) but there

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360067141

- [review_comment] 2025-09-18T16:02:35Z by coderabbitai[bot] (Makefile:6)

```text
In Makefile around line 6: the LDFLAGS uses -X main.version=$(VERSION) but there
is no package-level var named version in any package main; either add a
package-level variable declaration like `var version string` in your app's main
package (e.g., cmd/<app>/main.go) or change the -X value to the correct
fully-qualified import path and symbol that actually exists (e.g., -X
github.com/your/module/cmd/<app>.version=$(VERSION)); update the Makefile or the
main package accordingly so the linker symbol matches an existing variable.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In Makefile around lines 22 to 24, the run-tui target hardcodes "./bin/tui"

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360067151

- [review_comment] 2025-09-18T16:02:35Z by coderabbitai[bot] (Makefile:24)

```text
In Makefile around lines 22 to 24, the run-tui target hardcodes "./bin/tui"
instead of using the project variable; change the invocation to use
$(BIN_DIR)/tui (or the existing BIN_DIR variable) so the binary directory is
referenced consistently, and ensure any other references in this target (e.g.,
path to config) follow the same variable usage if applicable.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In create_postmortem_tasks.py around lines 18-19 (and also apply the same change

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360072243

- [review_comment] 2025-09-18T16:03:43Z by coderabbitai[bot] (create_postmortem_tasks.py:19)

```text
In create_postmortem_tasks.py around lines 18-19 (and also apply the same change
to lines 72-73), the timestamps currently include varying sub-second precision;
normalize them to seconds precision and a stable Z suffix by removing
microseconds and formatting the datetime in UTC with a trailing "Z". Update the
code that builds those "created_at" values to zero out microseconds (or
otherwise format to seconds) and emit an ISO-8601 string with a literal "Z"
timezone indicator so all timestamps are stable and consistent.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In create_postmortem_tasks.py around lines 132 to 141 the code writes JSON files

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360072263

- [review_comment] 2025-09-18T16:03:43Z by coderabbitai[bot] (create_postmortem_tasks.py:141)

```text
In create_postmortem_tasks.py around lines 132 to 141 the code writes JSON files
directly which can leave corrupted or partial files on crash; change the writes
to atomically replace the target: write to a temporary file in the same
directory (e.g., same filename + a .tmp suffix or use
tempfile.NamedTemporaryFile(dir=... , delete=False)), flush and fsync the file
descriptor to ensure data is on disk, close it, then call os.replace(temp_path,
final_path) to atomically move it into place; apply this for both the per-task
loop and the coordinator task write.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In create_postmortem_tasks.py around lines 134-141 (and similarly lines

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360072286

- [review_comment] 2025-09-18T16:03:43Z by coderabbitai[bot] (create_postmortem_tasks.py:141)

```text
In create_postmortem_tasks.py around lines 134-141 (and similarly lines
139-141), the JSON files are opened without specifying encoding and json.dump
defaults to ASCII-escaping non-ASCII chars; update both file writes to open(...,
'w', encoding='utf-8') and call json.dump with ensure_ascii=False and
deterministic options (e.g., sort_keys=True, keep indent) so output is UTF-8,
non-ASCII characters are preserved, and file contents are deterministic.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/admin-api/k8s-deployment.yaml around lines 41 to 44, the inline

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360072299

- [review_comment] 2025-09-18T16:03:44Z by coderabbitai[bot] (deployments/admin-api/k8s-deployment.yaml:44)

```text
In deployments/admin-api/k8s-deployment.yaml around lines 41 to 44, the inline
comments after the empty secret values do not have two spaces before the “# …”
which fails the linter; update each line so there are exactly two spaces between
the value and the inline comment (e.g., change '"" #' to '""  #' for both
jwt-secret and redis-password) and save the file.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In deployments/README.md around lines 186 to 190 there is a duplicated bullet

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360072323

- [review_comment] 2025-09-18T16:03:44Z by coderabbitai[bot] (deployments/README.md:190)

```text
In deployments/README.md around lines 186 to 190 there is a duplicated bullet
"Prometheus metrics: Detailed health metrics"; remove the duplicate so the list
contains only one entry for Prometheus metrics (leave the other bullets
unchanged) and ensure spacing/formatting of the remaining list item matches the
surrounding bullets.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/12_performance_baseline.md around lines 26 to 29, the Redis options (AOF

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360072335

- [review_comment] 2025-09-18T16:03:44Z by coderabbitai[bot] (docs/12_performance_baseline.md:29)

```text
In docs/12_performance_baseline.md around lines 26 to 29, the Redis options (AOF
disabled, noeviction, tcp-keepalive=60) are asserted but not actually applied by
the documented docker run; update the docs to show a docker run (or
docker-compose) invocation that explicitly passes those Redis configuration
options to the container (or mounts a redis.conf) so they are enforced
reproducibly — specifically ensure AOF is disabled (appendonly no),
maxmemory-policy is set to noeviction, and tcp-keepalive is set to 60 in the
command or config file referenced by the doc.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/12_performance_baseline.md around lines 58 to 62, the example uses

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360072364

- [review_comment] 2025-09-18T16:03:45Z by coderabbitai[bot] (docs/12_performance_baseline.md:62)

```text
In docs/12_performance_baseline.md around lines 58 to 62, the example uses
--bench-rate=1000 which conflates jobs/sec vs jobs/min and invalidates the "≥1k
jobs/min" claim; change the example to use a per‑second rate that matches the
target (e.g., --bench-rate=20 to target ≈1.2k/min), explicitly state the unit in
the flag description near the example ("--bench-rate is jobs/second"), and
update the Expected Results text to reflect the corrected rate/throughput
numbers.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/12_performance_baseline.md around lines 76 to 77, the expected-results

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360072373

- [review_comment] 2025-09-18T16:03:45Z by coderabbitai[bot] (docs/12_performance_baseline.md:77)

```text
In docs/12_performance_baseline.md around lines 76 to 77, the expected-results
sentence currently references bench-rate=1000 and a throughput/latency target
that contradicts the example bench command; update the text so the targets align
with the example using --bench-rate=20 (for example state the expected
throughput and p95 latency appropriate for bench-rate=20 on a 4 vCPU node),
replacing the numbers "bench-count=2000, bench-rate=1000 should achieve ≥1k
jobs/min throughput, with p95 latency < 2s" with phrasing that references
--bench-rate=20 and gives realistic throughput/latency targets for that rate.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/anomaly-radar-openapi.yaml around lines 32-35, the inline mapping

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360072387

- [review_comment] 2025-09-18T16:03:45Z by coderabbitai[bot] (docs/api/anomaly-radar-openapi.yaml:35)

```text
In docs/api/anomaly-radar-openapi.yaml around lines 32-35, the inline mapping
style like "{ $ref: '#/components/responses/Unauthorized' }" violates yamllint;
replace each inline curly-brace map with block-style YAML (use a named key
mapping, e.g. set the response code to a block mapping with $ref on its own
line) and apply the same conversion to the other reported ranges (51-53, 74-78,
116-119, 135-137, 153-156, 177-179, 200-203, 225-228) so all inline "{ $ref: ...
}" occurrences are converted to block mappings.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/anomaly-radar-openapi.yaml around lines 80 to 101 (and also apply

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360072403

- [review_comment] 2025-09-18T16:03:45Z by coderabbitai[bot] (docs/api/anomaly-radar-openapi.yaml:101)

```text
In docs/api/anomaly-radar-openapi.yaml around lines 80 to 101 (and also apply
the same change at lines 341-345), the query parameters "window" and
"max_samples" only document defaults in prose; update their parameter schemas to
include explicit default values: add default: "24h" under the window schema
(type: string) and default: 1000 under the max_samples schema (type: integer),
ensuring the OpenAPI spec reflects the defaults directly in the schema for both
occurrences.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/anomaly-radar-openapi.yaml around lines 229-236 (and also update the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360072412

- [review_comment] 2025-09-18T16:03:45Z by coderabbitai[bot] (docs/api/anomaly-radar-openapi.yaml:236)

```text
In docs/api/anomaly-radar-openapi.yaml around lines 229-236 (and also update the
corresponding block at lines 17-21), the OpenAPI components/responses section
currently omits the percentiles endpoint definition referenced in the Markdown
docs; add a complete path entry for GET /api/v1/anomaly-radar/percentiles
including its operationId, parameters, security, responses (200 with schema for
the percentiles payload, and relevant 4xx/5xx responses), and any referenced
component schemas, or alternatively add a reusable response/schema under
components and reference it from the path so the OpenAPI contract matches the
documented endpoint.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/anomaly-radar-openapi.yaml around lines 404-412, the 'metrics' array

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360072433

- [review_comment] 2025-09-18T16:03:46Z by coderabbitai[bot] (docs/api/anomaly-radar-openapi.yaml:412)

```text
In docs/api/anomaly-radar-openapi.yaml around lines 404-412, the 'metrics' array
(and other arrays) lack a maxItems constraint; update the schemas to add a
sensible maxItems value consistent with your API pagination/default limits
(e.g., default page size or a documented upper bound) to 'metrics' and any other
unbounded arrays in this file (notably the arrays at 450-458 and 120-137), and
propagate similar maxItems limits to any nested/ referenced array schemas (or
add global constants if you use reusable components) so OpenAPI validators no
longer flag unbounded arrays.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/anomaly-radar-slo-budget.md around lines 262 to 295 the example GET

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360072450

- [review_comment] 2025-09-18T16:03:46Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:295)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 262 to 295 the example GET
/config response includes extra fields "summary" and "is_valid" that are not
present in the OpenAPI GetConfigResponse; align the contract by either removing
"summary" and "is_valid" from this example in the docs or add those fields to
the OpenAPI GetConfigResponse schema and update the implementation to return
them (update schema, regenerate clients if any, and ensure server handler sets
these fields) so the documentation, schema, and code are consistent.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In docs/api/anomaly-radar-slo-budget.md around lines 582-592 (also apply changes

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360072469

- [review_comment] 2025-09-18T16:03:46Z by coderabbitai[bot] (docs/api/anomaly-radar-slo-budget.md:592)

```text
In docs/api/anomaly-radar-slo-budget.md around lines 582-592 (also apply changes
at 348-366 and 465-472), the "Key Metrics" list omits p90 whereas other payloads
include it; standardize percentile metrics to p50/p90/p95/p99 across the
document by adding p90 to the "Latency Percentiles" bullet and update any
example responses or payloads that currently list p50/p95/p99 to include p90 so
all endpoints and docs consistently use p50/p90/p95/p99.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD

## In README.md around lines 201 to 203, the docker compose command references the

- URL: https://github.com/flyingrobots/go-redis-work-queue/pull/3#discussion_r2360072483

- [review_comment] 2025-09-18T16:03:46Z by coderabbitai[bot] (README.md:203)

```text
In README.md around lines 201 to 203, the docker compose command references the
wrong path (deploy/docker-compose.yml); update the command to point to the
actual file location deployments/docker/docker-compose.yaml (e.g. docker compose
-f deployments/docker/docker-compose.yaml up --build) so the documented command
works for users.
```

> [!INFO]- **Pending**
> **Outcome**
> 
> | Accepted | Rejected | Confidence | Remarks |
> |----------|----------|------------|---------|
> | — | — | — | Pending review. |
>
> **What Did We Do About It?**
> TBD
>
> **Rationale**
> TBD
>
> **Alternatives Considered**
> TBD
>
> **Lesson(s) Learned**
> TBD
