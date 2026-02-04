import * as React from 'react';
import {
    Tbody,
    Tr,
    Td,
    ExpandableRowContent,
} from '@patternfly/react-table';
import {
    Label,
    TextContent,
    Text,
    TextVariants,
    Button,
    List,
    ListItem,
} from '@patternfly/react-core';
import {
    CheckCircleIcon,
    ExclamationTriangleIcon,
    ExclamationCircleIcon,
    InfoCircleIcon,
    ExternalLinkAltIcon,
} from '@patternfly/react-icons';
import { Finding } from '../types';
import { RemediationPanel } from './RemediationPanel';

interface FindingsTableRowProps {
    finding: Finding;
    rowIndex: number;
    isExpanded: boolean;
    onToggle: (finding: Finding, isExpanded: boolean) => void;
}

const getStatusIcon = (status: string) => {
    switch (status) {
        case 'PASS':
            return <CheckCircleIcon color="var(--pf-v5-global--success-color--100)" />;
        case 'WARN':
            return <ExclamationTriangleIcon color="var(--pf-v5-global--warning-color--100)" />;
        case 'FAIL':
            return <ExclamationCircleIcon color="var(--pf-v5-global--danger-color--100)" />;
        case 'INFO':
        default:
            return <InfoCircleIcon color="var(--pf-v5-global--info-color--100)" />;
    }
};

const getStatusLabel = (status: string) => {
    switch (status) {
        case 'PASS':
            return <Label color="green">{status}</Label>;
        case 'WARN':
            return <Label color="orange">{status}</Label>;
        case 'FAIL':
            return <Label color="red">{status}</Label>;
        case 'INFO':
        default:
            return <Label color="blue">{status}</Label>;
    }
};

const isValidUrl = (urlString: string) => {
    try {
        const url = new URL(urlString);
        return url.protocol === 'http:' || url.protocol === 'https:';
    } catch (e) {
        return false;
    }
};

// Memoized component to prevent re-rendering all rows when one is expanded/collapsed.
// This significantly improves performance for large lists of findings.
export const FindingsTableRow = React.memo(({ finding, rowIndex, isExpanded, onToggle }: FindingsTableRowProps) => {
    const handleToggle = () => {
        onToggle(finding, !isExpanded);
    };

    return (
        <Tbody isExpanded={isExpanded}>
            <Tr>
                <Td
                    expand={{
                        rowIndex,
                        isExpanded: isExpanded,
                        onToggle: handleToggle,
                    }}
                />
                <Td dataLabel="Status">
                    {getStatusIcon(finding.status)} {getStatusLabel(finding.status)}
                </Td>
                <Td dataLabel="Category"><Label>{finding.category}</Label></Td>
                <Td dataLabel="Finding">{finding.title}</Td>
                <Td dataLabel="Resource">
                    {finding.resource
                        ? `${finding.namespace ? `${finding.namespace}/` : ''}${finding.resource}`
                        : '-'}
                </Td>
            </Tr>
            <Tr isExpanded={isExpanded}>
                <Td colSpan={5}>
                    <ExpandableRowContent>
                        <TextContent>
                            <Text component={TextVariants.h4}>Description</Text>
                            <Text>{finding.description}</Text>
                            {finding.impact && (
                                <>
                                    <Text component={TextVariants.h4}>Impact</Text>
                                    <Text>{finding.impact}</Text>
                                </>
                            )}
                            {finding.recommendation && (
                                <>
                                    <Text component={TextVariants.h4}>Recommendation</Text>
                                    <Text>{finding.recommendation}</Text>
                                </>
                            )}
                            {finding.references && finding.references.length > 0 && (
                                <>
                                    <Text component={TextVariants.h4}>References</Text>
                                    <List>
                                        {finding.references.map((ref, i) => {
                                            if (isValidUrl(ref)) {
                                                return (
                                                    <ListItem key={i}>
                                                        <Button
                                                            variant="link"
                                                            isInline
                                                            component="a"
                                                            href={ref}
                                                            target="_blank"
                                                            rel="noopener noreferrer"
                                                            icon={<ExternalLinkAltIcon />}
                                                            iconPosition="end"
                                                            aria-label={`${ref} (opens in new tab)`}
                                                        >
                                                            {ref}
                                                        </Button>
                                                    </ListItem>
                                                );
                                            }
                                            return (
                                                <ListItem key={i}>
                                                    {ref}
                                                </ListItem>
                                            );
                                        })}
                                    </List>
                                </>
                            )}
                        </TextContent>
                        {finding.remediation && (
                            <RemediationPanel remediation={finding.remediation} />
                        )}
                    </ExpandableRowContent>
                </Td>
            </Tr>
        </Tbody>
    );
});

FindingsTableRow.displayName = 'FindingsTableRow';
