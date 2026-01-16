# Architecture Documentation

This document provides visual diagrams to help understand the internal architecture and workflows of the Cluster Assessment Operator.

## High-Level Architecture

```mermaid
flowchart TB
    subgraph Kubernetes["OpenShift Cluster"]
        CR["ClusterAssessment CR"]
        Controller["Assessment Controller"]
        Registry["Validator Registry"]
        
        subgraph Validators["12 Validators"]
            direction LR
            V1["version"]
            V2["nodes"]
            V3["machineconfig"]
            V4["apiserver"]
            V5["operators"]
            V6["certificates"]
            V7["etcdbackup"]
            V8["security"]
            V9["networking"]
            V10["storage"]
            V11["monitoring"]
            V12["deprecation"]
        end
        
        Runner["Validator Runner"]
        Reporter["Report Generator"]
        ConfigMap["ConfigMap\n(JSON/HTML/PDF)"]
        Metrics["Prometheus Metrics"]
    end
    
    User["Platform Engineer"] --> CR
    CR --> Controller
    Controller --> Runner
    Runner --> Registry
    Registry --> Validators
    Validators -->|Findings| Runner
    Runner -->|All Findings| Controller
    Controller --> Reporter
    Reporter --> ConfigMap
    Controller --> Metrics
    ConfigMap --> User
    Metrics --> Prometheus["Prometheus/Alertmanager"]
```

## Component Interaction

```mermaid
sequenceDiagram
    participant User
    participant CR as ClusterAssessment CR
    participant Controller as Assessment Controller
    participant Runner as Validator Runner
    participant Registry as Validator Registry
    participant Validators
    participant K8s as Kubernetes API
    participant Reporter as Report Generator
    participant CM as ConfigMap

    User->>CR: Create/Update
    CR->>Controller: Reconcile Event
    Controller->>Controller: Check Schedule/Phase
    
    alt Scheduled Assessment
        Controller->>Controller: Calculate Next Run Time
    end
    
    Controller->>Runner: Run(profile, validators[])
    Runner->>Registry: Get Validators
    Registry-->>Runner: []Validator
    
    loop For Each Validator
        Runner->>Validators: Validate(ctx, client, profile)
        Validators->>K8s: Read-only API calls
        K8s-->>Validators: Cluster Resources
        Validators-->>Runner: []Finding
    end
    
    Runner-->>Controller: All Findings
    Controller->>Controller: Filter by MinSeverity
    Controller->>Controller: Calculate Summary & Score
    Controller->>Reporter: Generate Reports
    Reporter-->>Controller: JSON, HTML, PDF
    Controller->>CM: Store Report
    Controller->>CR: Update Status
    CR-->>User: View Results
```

## Validator Categories

```mermaid
mindmap
  root((Validators))
    Platform
      version
        OpenShift version
        Update channel
        Available updates
      nodes
        Node count
        Conditions
        Role distribution
      machineconfig
        MCP health
        Custom configs
      apiserver
        API status
        etcd health
        Encryption
        Audit logging
      operators
        CSV states
        ClusterOperator health
      etcdbackup
        OADP/Velero
        Backup CronJobs
    Security
      certificates
        TLS expiration
        Custom certs
      security
        Cluster-admin bindings
        Privileged pods
        RBAC audit
    Networking
      networking
        CNI type
        NetworkPolicies
        Ingress config
    Storage
      storage
        StorageClasses
        Default SC
        CSI drivers
    Observability
      monitoring
        Cluster monitoring
        User workload monitoring
    Compatibility
      deprecation
        Deprecated patterns
        Missing probes
```

## Assessment Lifecycle

```mermaid
stateDiagram-v2
    [*] --> Pending: CR Created
    
    Pending --> Running: Controller Picks Up
    
    Running --> Completed: All Validators Pass
    Running --> Failed: Critical Error
    
    Completed --> Running: Schedule Triggers
    Failed --> Running: Schedule Triggers
    
    Completed --> [*]: CR Deleted
    Failed --> [*]: CR Deleted
    
    note right of Pending
        Waiting for controller
        to process
    end note
    
    note right of Running
        Validators executing
        read-only checks
    end note
    
    note right of Completed
        Findings stored in
        status and ConfigMap
    end note
```

## Profile Comparison

```mermaid
graph LR
    subgraph Production["Production Profile (Strict)"]
        P1["Min 3 Control Plane Nodes"]
        P2["Min 3 Worker Nodes"]
        P3["NetworkPolicies Required"]
        P4["Privileged Containers Blocked"]
        P5["Max Update Age: 90 days"]
    end
    
    subgraph Development["Development Profile (Relaxed)"]
        D1["Min 1 Control Plane Node"]
        D2["Min 1 Worker Node"]
        D3["NetworkPolicies Optional"]
        D4["Privileged Containers Allowed"]
        D5["Max Update Age: 180 days"]
    end
    
    style Production fill:#e74c3c,color:#fff
    style Development fill:#27ae60,color:#fff
```

## Finding Severity Flow

```mermaid
flowchart LR
    subgraph Validators
        Check["Validation Check"]
    end
    
    Check --> Decision{Result?}
    
    Decision -->|All Good| PASS["✅ PASS"]
    Decision -->|Observation| INFO["ℹ️ INFO"]
    Decision -->|Review Needed| WARN["⚠️ WARN"]
    Decision -->|Action Required| FAIL["❌ FAIL"]
    
    subgraph Filtering["MinSeverity Filter"]
        PASS --> Filter
        INFO --> Filter
        WARN --> Filter
        FAIL --> Filter
    end
    
    Filter -->|"minSeverity: WARN"| Output["Final Report"]
    
    style PASS fill:#27ae60,color:#fff
    style INFO fill:#3498db,color:#fff
    style WARN fill:#f39c12,color:#fff
    style FAIL fill:#e74c3c,color:#fff
```

## Report Storage Options

```mermaid
flowchart TB
    Assessment["ClusterAssessment"]
    
    Assessment --> Storage{reportStorage}
    
    Storage --> ConfigMap["ConfigMap Storage"]
    Storage --> Git["Git Export (Future)"]
    
    subgraph ConfigMapDetails["ConfigMap Options"]
        CM_JSON["report.json"]
        CM_HTML["report.html"]
        CM_PDF["report.pdf"]
    end
    
    ConfigMap --> ConfigMapDetails
    
    subgraph GitDetails["Git Options"]
        G_URL["Repository URL"]
        G_Branch["Branch"]
        G_Path["Directory Path"]
        G_Secret["Credentials Secret"]
    end
    
    Git --> GitDetails
```

## Metrics Architecture

```mermaid
flowchart LR
    subgraph Operator["Assessment Operator"]
        Controller["Controller"]
        MetricsEndpoint["/metrics endpoint"]
    end
    
    Controller -->|Records| MetricsEndpoint
    
    subgraph Metrics["Exported Metrics"]
        M1["cluster_assessment_score"]
        M2["cluster_assessment_findings_total"]
        M3["cluster_assessment_findings_by_category"]
        M4["cluster_assessment_last_run_timestamp"]
        M5["cluster_assessment_duration_seconds"]
    end
    
    MetricsEndpoint --> Metrics
    
    Prometheus["Prometheus"] -->|Scrapes| MetricsEndpoint
    
    Prometheus --> Alertmanager["Alertmanager"]
    Prometheus --> Grafana["Grafana Dashboard"]
    
    Alertmanager --> Slack["Slack/PagerDuty"]
```

## Data Model

```mermaid
erDiagram
    ClusterAssessment ||--|| ClusterAssessmentSpec : has
    ClusterAssessment ||--|| ClusterAssessmentStatus : has
    
    ClusterAssessmentSpec ||--o| ReportStorageSpec : contains
    ReportStorageSpec ||--o| ConfigMapStorageSpec : has
    ReportStorageSpec ||--o| GitStorageSpec : has
    
    ClusterAssessmentStatus ||--|| ClusterInfo : contains
    ClusterAssessmentStatus ||--|| AssessmentSummary : contains
    ClusterAssessmentStatus ||--o{ Finding : contains
    ClusterAssessmentStatus ||--o{ Condition : contains
    
    ClusterAssessmentSpec {
        string schedule
        string profile
        array validators
        bool suspend
        string minSeverity
    }
    
    ClusterInfo {
        string clusterID
        string clusterVersion
        string platform
        string channel
        int nodeCount
    }
    
    AssessmentSummary {
        int totalChecks
        int passCount
        int warnCount
        int failCount
        int infoCount
        int score
    }
    
    Finding {
        string id
        string validator
        string category
        string resource
        string status
        string title
        string description
        string recommendation
    }
```

## Deployment Architecture

```mermaid
flowchart TB
    subgraph OpenShift["OpenShift Cluster"]
        subgraph OperatorNS["cluster-assessment-operator namespace"]
            Deployment["Operator Deployment"]
            Pod["Manager Pod"]
            SA["ServiceAccount"]
        end
        
        subgraph RBAC["Cluster RBAC"]
            CR_Role["ClusterRole\n(read-only)"]
            CR_Binding["ClusterRoleBinding"]
        end
        
        subgraph CRD["Custom Resource Definitions"]
            CRD_CA["ClusterAssessment CRD"]
        end
        
        subgraph UserNS["Any Namespace"]
            Assessment["ClusterAssessment CR"]
            Report["Report ConfigMap"]
        end
    end
    
    Deployment --> Pod
    Pod --> SA
    SA --> CR_Binding
    CR_Binding --> CR_Role
    CRD_CA --> Assessment
    Pod -->|Reconciles| Assessment
    Pod -->|Creates| Report
    
    style Pod fill:#3498db,color:#fff
    style Assessment fill:#9b59b6,color:#fff
```
