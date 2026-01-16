// Shared TypeScript types for the cluster assessment plugin

export interface ClusterAssessment {
    metadata: {
        name: string;
        creationTimestamp: string;
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
}
