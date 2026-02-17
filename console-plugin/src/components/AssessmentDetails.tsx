import * as React from 'react';
import { useParams } from 'react-router-dom';
import {
    Page,
    PageSection,
    Title,
    Breadcrumb,
    BreadcrumbItem,
    Split,
    SplitItem,
    Label,
    Flex,
    FlexItem,
    EmptyState,
    EmptyStateBody,
    EmptyStateIcon,
    Spinner,
    Tabs,
    Tab,
    TabTitleText,
    Button,
    Alert,
} from '@patternfly/react-core';
import {
    ExclamationCircleIcon,
    CheckCircleIcon,
    ExclamationTriangleIcon,
} from '@patternfly/react-icons';
import { useK8sWatchResource, k8sPatch } from '@openshift-console/dynamic-plugin-sdk';
import { Link } from 'react-router-dom';
import { ClusterAssessment } from '../types';
import { ScoreGauge } from './ScoreGauge';
import { FindingsTable } from './FindingsTable';
import DeltaBanner from './DeltaBanner';
import TrendChart from './TrendChart';
import './styles.css';

const clusterAssessmentResource = (name: string) => ({
    groupVersionKind: {
        group: 'assessment.openshift.io',
        version: 'v1alpha1',
        kind: 'ClusterAssessment',
    },
    name,
    isList: false,
});

export default function AssessmentDetails() {
    const { name } = useParams<{ name: string }>();
    const [activeTabKey, setActiveTabKey] = React.useState<string | number>(0);
    const [rerunStatus, setRerunStatus] = React.useState<'idle' | 'running' | 'success' | 'error'>('idle');
    const [rerunError, setRerunError] = React.useState<string>('');

    const handleRerun = async () => {
        if (!name) return;
        setRerunStatus('running');
        setRerunError('');
        try {
            await k8sPatch({
                model: {
                    apiGroup: 'assessment.openshift.io',
                    apiVersion: 'v1alpha1',
                    kind: 'ClusterAssessment',
                    plural: 'clusterassessments',
                    abbr: 'CA',
                    label: 'ClusterAssessment',
                    labelPlural: 'ClusterAssessments',
                },
                resource: { metadata: { name } },
                data: [
                    {
                        op: 'add',
                        path: '/metadata/annotations/assessment.openshift.io~1trigger',
                        value: 'run',
                    },
                ],
            });
            setRerunStatus('success');
            setTimeout(() => setRerunStatus('idle'), 3000);
        } catch (err) {
            setRerunStatus('error');
            setRerunError(String(err));
        }
    };

    // Stable assessment data that only gets replaced with valid complete data.
    // This prevents the UI from going blank when the watch returns empty/stale data.
    const [stableAssessment, setStableAssessment] = React.useState<ClusterAssessment | undefined>();

    const resourceConfig = React.useMemo(
        () => clusterAssessmentResource(name || ''),
        [name]
    );

    const [watchData, loaded, error] = useK8sWatchResource<ClusterAssessment>(resourceConfig);

    // Only update stableAssessment when the watch returns data with findings
    React.useEffect(() => {
        if (
            watchData &&
            watchData.status &&
            watchData.status.findings &&
            watchData.status.findings.length > 0
        ) {
            setStableAssessment(watchData);
        }
    }, [watchData]);

    // Use stableAssessment if available, otherwise fall back to watchData
    const assessment = stableAssessment || watchData;

    if (error && !stableAssessment) {
        return (
            <Page>
                <PageSection>
                    <EmptyState>
                        <EmptyStateIcon icon={ExclamationCircleIcon} />
                        <Title headingLevel="h4" size="lg">Error loading assessment</Title>
                        <EmptyStateBody>{String(error)}</EmptyStateBody>
                    </EmptyState>
                </PageSection>
            </Page>
        );
    }

    if (!loaded && !stableAssessment) {
        return (
            <Page>
                <PageSection>
                    <EmptyState>
                        <Spinner size="xl" />
                        <Title headingLevel="h4" size="lg">Loading assessment...</Title>
                    </EmptyState>
                </PageSection>
            </Page>
        );
    }

    const summary = assessment?.status?.summary;
    const clusterInfo = assessment?.status?.clusterInfo;
    const findings = assessment?.status?.findings || [];

    const getScoreClass = (score: number) => {
        if (score >= 80) return 'ca-plugin__score-value--good';
        if (score >= 60) return 'ca-plugin__score-value--warning';
        return 'ca-plugin__score-value--critical';
    };

    const score = summary?.score ?? 0;

    return (
        <Page>
            {/* Breadcrumb */}
            <PageSection variant="light">
                <Breadcrumb>
                    <BreadcrumbItem>
                        <Link to="/cluster-assessment">Cluster Assessment</Link>
                    </BreadcrumbItem>
                    <BreadcrumbItem isActive>{name}</BreadcrumbItem>
                </Breadcrumb>
            </PageSection>

            {/* Page Header */}
            <PageSection variant="light">
                <Split hasGutter>
                    <SplitItem isFilled>
                        <Flex spaceItems={{ default: 'spaceItemsMd' }} alignItems={{ default: 'alignItemsCenter' }}>
                            <FlexItem>
                                <Title headingLevel="h1">{name}</Title>
                            </FlexItem>
                            <FlexItem>
                                {assessment?.status?.phase === 'Completed' && (
                                    <Label color="green">Completed</Label>
                                )}
                                {assessment?.status?.phase === 'Running' && (
                                    <Label color="blue">Running</Label>
                                )}
                                {assessment?.status?.phase === 'Failed' && (
                                    <Label color="red">Failed</Label>
                                )}
                                {assessment?.status?.phase === 'Pending' && (
                                    <Label color="grey">Pending</Label>
                                )}
                            </FlexItem>
                        </Flex>
                    </SplitItem>
                    <SplitItem>
                        <Button
                            variant="secondary"
                            onClick={handleRerun}
                            isLoading={rerunStatus === 'running'}
                            isDisabled={rerunStatus === 'running' || assessment?.status?.phase === 'Running'}
                        >
                            {rerunStatus === 'success' ? 'Re-run Triggered!' : 'Re-run Assessment'}
                        </Button>
                    </SplitItem>
                </Split>
                {rerunStatus === 'error' && (
                    <Alert variant="danger" title="Re-run failed" isInline style={{ marginTop: '8px' }}>
                        {rerunError}
                    </Alert>
                )}
            </PageSection>

            {/* Cards Grid Section */}
            <PageSection>
                <div className="ca-plugin__details-grid">
                    {/* Health Score Card */}
                    <div className="ca-plugin__details-card ca-plugin__score-card">
                        <div className="ca-plugin__details-card-header">Health Score</div>
                        <div className="ca-plugin__details-card-body" style={{ textAlign: 'center' }}>
                            <div className="ca-plugin__score-gauge">
                                <ScoreGauge score={score} />
                            </div>
                            <div className={`ca-plugin__score-value ${getScoreClass(score)}`}>
                                {score}%
                            </div>
                            <div className="ca-plugin__score-label">
                                {score >= 80 ? 'Good' : score >= 60 ? 'Warning' : 'Critical'}
                            </div>
                        </div>
                    </div>

                    {/* Cluster Info Card */}
                    <div className="ca-plugin__details-card">
                        <div className="ca-plugin__details-card-header">Cluster Info</div>
                        <div className="ca-plugin__details-card-body">
                            <ul className="ca-plugin__info-list">
                                <li className="ca-plugin__info-item">
                                    <span className="ca-plugin__info-label">Version</span>
                                    <span className="ca-plugin__info-value">{clusterInfo?.clusterVersion ?? 'N/A'}</span>
                                </li>
                                <li className="ca-plugin__info-item">
                                    <span className="ca-plugin__info-label">Platform</span>
                                    <span className="ca-plugin__info-value">{clusterInfo?.platform ?? 'N/A'}</span>
                                </li>
                                <li className="ca-plugin__info-item">
                                    <span className="ca-plugin__info-label">Nodes</span>
                                    <span className="ca-plugin__info-value">{clusterInfo?.nodeCount ?? 'N/A'}</span>
                                </li>
                            </ul>
                        </div>
                    </div>

                    {/* Configuration Card */}
                    <div className="ca-plugin__details-card">
                        <div className="ca-plugin__details-card-header">Configuration</div>
                        <div className="ca-plugin__details-card-body">
                            <ul className="ca-plugin__info-list">
                                <li className="ca-plugin__info-item">
                                    <span className="ca-plugin__info-label">Profile</span>
                                    <span className="ca-plugin__info-value">
                                        <Label color={assessment?.spec?.profile === 'production' ? 'blue' : 'green'}>
                                            {assessment?.spec?.profile ?? 'production'}
                                        </Label>
                                    </span>
                                </li>
                                <li className="ca-plugin__info-item">
                                    <span className="ca-plugin__info-label">Schedule</span>
                                    <span className="ca-plugin__info-value">{assessment?.spec?.schedule || 'One-time'}</span>
                                </li>
                                <li className="ca-plugin__info-item">
                                    <span className="ca-plugin__info-label">Last Run</span>
                                    <span className="ca-plugin__info-value">
                                        {assessment?.status?.lastRunTime
                                            ? new Date(assessment.status.lastRunTime).toLocaleString()
                                            : 'Never'}
                                    </span>
                                </li>
                            </ul>
                        </div>
                    </div>

                    {/* Results Summary Card */}
                    <div className="ca-plugin__details-card">
                        <div className="ca-plugin__details-card-header">Results Summary</div>
                        <div className="ca-plugin__details-card-body">
                            <div className="ca-plugin__result-item">
                                <span className="ca-plugin__result-label">Total Checks</span>
                                <span className="ca-plugin__result-value ca-plugin__result-value--total">{summary?.totalChecks ?? 0}</span>
                            </div>
                            <div className="ca-plugin__result-item">
                                <span className="ca-plugin__result-label">
                                    <CheckCircleIcon className="ca-plugin__status--pass" style={{ marginRight: '8px' }} />
                                    Passed
                                </span>
                                <span className="ca-plugin__result-value ca-plugin__result-value--pass">{summary?.passCount ?? 0}</span>
                            </div>
                            <div className="ca-plugin__result-item">
                                <span className="ca-plugin__result-label">
                                    <ExclamationTriangleIcon className="ca-plugin__status--warn" style={{ marginRight: '8px' }} />
                                    Warnings
                                </span>
                                <span className="ca-plugin__result-value ca-plugin__result-value--warn">{summary?.warnCount ?? 0}</span>
                            </div>
                            <div className="ca-plugin__result-item">
                                <span className="ca-plugin__result-label">
                                    <ExclamationCircleIcon className="ca-plugin__status--fail" style={{ marginRight: '8px' }} />
                                    Failed
                                </span>
                                <span className="ca-plugin__result-value ca-plugin__result-value--fail">{summary?.failCount ?? 0}</span>
                            </div>
                        </div>
                    </div>
                </div>

                {/* Delta Banner */}
                <DeltaBanner delta={assessment?.status?.delta} />

                {/* Findings Table */}
                <div className="ca-plugin__table-card">
                    <Tabs
                        activeKey={activeTabKey}
                        onSelect={(_event, tabIndex) => setActiveTabKey(tabIndex)}
                        style={{ padding: '0 16px' }}
                    >
                        <Tab eventKey={0} title={<TabTitleText>All Findings ({findings.length})</TabTitleText>}>
                            <div style={{ padding: '16px' }}>
                                <FindingsTable findings={findings} />
                            </div>
                        </Tab>
                        <Tab
                            eventKey={1}
                            title={
                                <TabTitleText>
                                    Issues ({findings.filter((f) => f.status === 'FAIL' || f.status === 'WARN').length})
                                </TabTitleText>
                            }
                        >
                            <div style={{ padding: '16px' }}>
                                <FindingsTable
                                    findings={findings.filter((f) => f.status === 'FAIL' || f.status === 'WARN')}
                                />
                            </div>
                        </Tab>
                        <Tab eventKey={2} title={<TabTitleText>History &amp; Trends</TabTitleText>}>
                            <div style={{ padding: '16px' }}>
                                <TrendChart assessmentName={name || ''} />
                            </div>
                        </Tab>
                    </Tabs>
                </div>
            </PageSection>
        </Page>
    );
}
