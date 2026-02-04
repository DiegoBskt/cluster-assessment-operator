import * as React from 'react';
import {
    Card,
    CardBody,
    CardTitle,
    EmptyState,
    EmptyStateBody,
    EmptyStateIcon,
    Title,
} from '@patternfly/react-core';
import { TrendUpIcon } from '@patternfly/react-icons';
import { useK8sWatchResource } from '@openshift-console/dynamic-plugin-sdk';

interface AssessmentSnapshot {
    metadata: {
        name: string;
        creationTimestamp: string;
    };
    spec: {
        assessmentName: string;
        profile: string;
    };
    status?: {
        runTime?: string;
        summary?: {
            score?: number;
            passCount: number;
            warnCount: number;
            failCount: number;
            infoCount: number;
        };
    };
}

interface TrendChartProps {
    assessmentName: string;
}

const snapshotResource = (assessmentName: string) => ({
    groupVersionKind: {
        group: 'assessment.openshift.io',
        version: 'v1alpha1',
        kind: 'AssessmentSnapshot',
    },
    isList: true,
    namespaced: false,
    selector: {
        matchLabels: {
            'assessment.openshift.io/name': assessmentName,
        },
    },
});

export default function TrendChart({ assessmentName }: TrendChartProps) {
    const [snapshots, loaded, error] = useK8sWatchResource<AssessmentSnapshot[]>(
        snapshotResource(assessmentName),
    );

    if (!loaded) {
        return (
            <Card>
                <CardBody>Loading trend data...</CardBody>
            </Card>
        );
    }

    if (error) {
        return (
            <Card>
                <CardBody>Error loading trend data: {String(error)}</CardBody>
            </Card>
        );
    }

    const sorted = [...(snapshots || [])]
        .filter((s) => s.status?.runTime && s.status?.summary?.score !== undefined)
        .sort((a, b) => {
            const timeA = new Date(a.status!.runTime!).getTime();
            const timeB = new Date(b.status!.runTime!).getTime();
            return timeA - timeB;
        });

    if (sorted.length === 0) {
        return (
            <Card>
                <CardBody>
                    <EmptyState>
                        <EmptyStateIcon icon={TrendUpIcon} />
                        <Title headingLevel="h4" size="lg">No History Available</Title>
                        <EmptyStateBody>
                            Trend data will appear after multiple assessment runs.
                        </EmptyStateBody>
                    </EmptyState>
                </CardBody>
            </Card>
        );
    }

    return (
        <Card>
            <CardTitle>Score History ({sorted.length} snapshots)</CardTitle>
            <CardBody>
                <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                    <thead>
                        <tr style={{ borderBottom: '2px solid var(--pf-v5-global--BorderColor--100)' }}>
                            <th style={{ padding: '8px', textAlign: 'left' }}>Date</th>
                            <th style={{ padding: '8px', textAlign: 'center' }}>Score</th>
                            <th style={{ padding: '8px', textAlign: 'center' }}>Pass</th>
                            <th style={{ padding: '8px', textAlign: 'center' }}>Warn</th>
                            <th style={{ padding: '8px', textAlign: 'center' }}>Fail</th>
                        </tr>
                    </thead>
                    <tbody>
                        {sorted.map((snapshot) => {
                            const summary = snapshot.status!.summary!;
                            const score = summary.score ?? 0;
                            const scoreColor =
                                score >= 80
                                    ? 'var(--pf-v5-global--success-color--100)'
                                    : score >= 50
                                    ? 'var(--pf-v5-global--warning-color--100)'
                                    : 'var(--pf-v5-global--danger-color--100)';

                            return (
                                <tr
                                    key={snapshot.metadata.name}
                                    style={{ borderBottom: '1px solid var(--pf-v5-global--BorderColor--100)' }}
                                >
                                    <td style={{ padding: '8px' }}>
                                        {new Date(snapshot.status!.runTime!).toLocaleString()}
                                    </td>
                                    <td
                                        style={{
                                            padding: '8px',
                                            textAlign: 'center',
                                            fontWeight: 'bold',
                                            color: scoreColor,
                                        }}
                                    >
                                        {score}
                                    </td>
                                    <td style={{ padding: '8px', textAlign: 'center' }}>
                                        {summary.passCount}
                                    </td>
                                    <td style={{ padding: '8px', textAlign: 'center' }}>
                                        {summary.warnCount}
                                    </td>
                                    <td style={{ padding: '8px', textAlign: 'center' }}>
                                        {summary.failCount}
                                    </td>
                                </tr>
                            );
                        })}
                    </tbody>
                </table>
            </CardBody>
        </Card>
    );
}
