// Shared TypeScript types for the cluster assessment plugin

export interface ClusterAssessment {
    metadata: {
        name: string;
        creationTimestamp: string;
        annotations?: Record<string, string>;
    };
    spec: {
        profile?: string;
        schedule?: string;
    };
    status?: {
        phase?: string;
        lastRunTime?: string;
        summary?: {
            score?: number;
            passCount: number;
            warnCount: number;
            failCount: number;
            infoCount: number;
            totalChecks: number;
        };
        clusterInfo?: {
            clusterVersion?: string;
            platform?: string;
            nodeCount?: number;
        };
        findings?: Finding[];
        delta?: DeltaSummary;
        snapshotCount?: number;
    };
}

export interface Finding {
    id: string;
    validator: string;
    category: string;
    resource?: string;
    namespace?: string;
    status: 'PASS' | 'WARN' | 'FAIL' | 'INFO';
    title: string;
    description: string;
    impact?: string;
    recommendation?: string;
    references?: string[];
    remediation?: RemediationGuidance;
}

export type RemediationSafety = 'safe-apply' | 'requires-review' | 'destructive';

export interface RemediationCommand {
    command: string;
    description?: string;
    requiresConfirmation?: boolean;
}

export interface RemediationGuidance {
    safety: RemediationSafety;
    commands?: RemediationCommand[];
    documentationURL?: string;
    estimatedImpact?: string;
    prerequisites?: string[];
}

export interface DeltaSummary {
    newFindings?: string[];
    resolvedFindings?: string[];
    regressionFindings?: string[];
    improvedFindings?: string[];
    scoreDelta?: number;
}

export interface AssessmentProfile {
    metadata: {
        name: string;
        creationTimestamp: string;
    };
    spec: {
        description?: string;
        basedOn?: string;
        thresholds?: Partial<ThresholdOverrides>;
        enabledValidators?: string[];
        disabledValidators?: string[];
        disabledChecks?: string[];
    };
    status?: {
        ready?: boolean;
        message?: string;
        resolvedValidatorCount?: number;
    };
}

export interface ThresholdOverrides {
    minControlPlaneNodes: number;
    minWorkerNodes: number;
    maxPodsPerNode: number;
    maxClusterAdminBindings: number;
    requireNetworkPolicy: boolean;
    requireResourceQuotas: boolean;
    requireLimitRanges: boolean;
    maxDaysWithoutUpdate: number;
    allowPrivilegedContainers: boolean;
    requireDefaultStorageClass: boolean;
}
