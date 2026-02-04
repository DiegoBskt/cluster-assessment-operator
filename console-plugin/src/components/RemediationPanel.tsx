import * as React from 'react';
import {
    Label,
    ClipboardCopy,
    TextContent,
    Text,
    TextVariants,
    List,
    ListItem,
    Button,
    Flex,
    FlexItem,
} from '@patternfly/react-core';
import {
    ExclamationTriangleIcon,
    ExternalLinkAltIcon,
} from '@patternfly/react-icons';
import { RemediationGuidance } from '../types';
import './styles.css';

interface RemediationPanelProps {
    remediation: RemediationGuidance;
}

const getSafetyLabel = (safety: string) => {
    switch (safety) {
        case 'safe-apply':
            return <Label color="green">Safe to Apply</Label>;
        case 'requires-review':
            return <Label color="orange">Requires Review</Label>;
        case 'destructive':
            return <Label color="red">Destructive</Label>;
        default:
            return <Label>{safety}</Label>;
    }
};

export const RemediationPanel: React.FC<RemediationPanelProps> = ({ remediation }) => {
    return (
        <div className="ca-plugin__remediation-panel">
            <TextContent>
                <Text component={TextVariants.h4}>
                    Remediation {getSafetyLabel(remediation.safety)}
                </Text>
            </TextContent>

            {remediation.estimatedImpact && (
                <div className="ca-plugin__remediation-impact">
                    <Text component={TextVariants.small}>
                        <strong>Impact:</strong> {remediation.estimatedImpact}
                    </Text>
                </div>
            )}

            {remediation.prerequisites && remediation.prerequisites.length > 0 && (
                <div className="ca-plugin__remediation-prereqs">
                    <Text component={TextVariants.small}><strong>Prerequisites:</strong></Text>
                    <List>
                        {remediation.prerequisites.map((prereq, i) => (
                            <ListItem key={i}>{prereq}</ListItem>
                        ))}
                    </List>
                </div>
            )}

            {remediation.commands && remediation.commands.length > 0 && (
                <div className="ca-plugin__remediation-commands">
                    {remediation.commands.map((cmd, i) => (
                        <div key={i} className="ca-plugin__remediation-command">
                            {cmd.description && (
                                <Flex spaceItems={{ default: 'spaceItemsSm' }} alignItems={{ default: 'alignItemsCenter' }}>
                                    <FlexItem>
                                        <Text component={TextVariants.small}>
                                            {cmd.requiresConfirmation && (
                                                <ExclamationTriangleIcon
                                                    color="var(--pf-v5-global--warning-color--100)"
                                                    style={{ marginRight: '4px' }}
                                                />
                                            )}
                                            {cmd.description}
                                        </Text>
                                    </FlexItem>
                                </Flex>
                            )}
                            <ClipboardCopy
                                isReadOnly
                                hoverTip="Copy"
                                clickTip="Copied"
                                variant="expansion"
                            >
                                {cmd.command}
                            </ClipboardCopy>
                        </div>
                    ))}
                </div>
            )}

            {remediation.documentationURL && (
                <div className="ca-plugin__remediation-docs">
                    <Button
                        variant="link"
                        isInline
                        component="a"
                        href={remediation.documentationURL}
                        target="_blank"
                        rel="noopener noreferrer"
                        icon={<ExternalLinkAltIcon />}
                        iconPosition="end"
                    >
                        Documentation
                    </Button>
                </div>
            )}
        </div>
    );
};

export default RemediationPanel;
