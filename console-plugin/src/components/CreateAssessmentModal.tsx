import * as React from 'react';
import {
    Modal,
    ModalVariant,
    Button,
    Form,
    FormGroup,
    TextInput,
    FormSelect,
    FormSelectOption,
    Checkbox,
    Alert,
    FormHelperText,
    HelperText,
    HelperTextItem,
} from '@patternfly/react-core';
import { ExclamationCircleIcon } from '@patternfly/react-icons';
import { k8sCreate, K8sModel, useK8sWatchResource } from '@openshift-console/dynamic-plugin-sdk';
import { AssessmentProfile } from '../types';

interface CreateAssessmentModalProps {
    isOpen: boolean;
    onClose: () => void;
    onCreated: () => void;
}

const clusterAssessmentModel: K8sModel = {
    apiVersion: 'v1alpha1',
    apiGroup: 'assessment.openshift.io',
    kind: 'ClusterAssessment',
    plural: 'clusterassessments',
    abbr: 'CA',
    label: 'Cluster Assessment',
    labelPlural: 'Cluster Assessments',
    namespaced: false,
};

const assessmentProfileResource = {
    groupVersionKind: {
        group: 'assessment.openshift.io',
        version: 'v1alpha1',
        kind: 'AssessmentProfile',
    },
    isList: true,
    namespaced: false,
};

const builtInProfiles = [
    { value: 'production', label: 'Production (Strict)', description: 'Strict checks suitable for production environments.' },
    { value: 'development', label: 'Development (Relaxed)', description: 'Relaxed checks suitable for development or test environments.' },
];

export default function CreateAssessmentModal({
    isOpen,
    onClose,
    onCreated,
}: CreateAssessmentModalProps) {
    const [name, setName] = React.useState('');
    const [profile, setProfile] = React.useState('production');
    const [enableHtml, setEnableHtml] = React.useState(true);
    const [enableJson, setEnableJson] = React.useState(true);
    const [enablePdf, setEnablePdf] = React.useState(false);
    const [isSubmitting, setIsSubmitting] = React.useState(false);
    const [error, setError] = React.useState<string | null>(null);

    const [customProfiles] = useK8sWatchResource<AssessmentProfile[]>(assessmentProfileResource);

    const profileOptions = React.useMemo(() => {
        const options = [...builtInProfiles];
        if (customProfiles && customProfiles.length > 0) {
            for (const cp of customProfiles) {
                if (cp.status?.ready !== false) {
                    options.push({
                        value: cp.metadata.name,
                        label: `${cp.metadata.name} (Custom - based on ${cp.spec.basedOn || 'production'})`,
                        description: cp.spec.description || `Custom profile based on ${cp.spec.basedOn || 'production'}.`,
                    });
                }
            }
        }
        return options;
    }, [customProfiles]);

    const selectedProfileDescription = React.useMemo(() => {
        const found = profileOptions.find(p => p.value === profile);
        return found?.description || '';
    }, [profile, profileOptions]);

    const handleSubmit = async () => {
        if (!name.trim()) {
            setError('Name is required');
            return;
        }

        if (!enableHtml && !enableJson && !enablePdf) {
            setError('At least one report format is required');
            return;
        }

        setIsSubmitting(true);
        setError(null);

        const formats: string[] = [];
        if (enableJson) formats.push('json');
        if (enableHtml) formats.push('html');
        if (enablePdf) formats.push('pdf');

        const resource = {
            apiVersion: 'assessment.openshift.io/v1alpha1',
            kind: 'ClusterAssessment',
            metadata: {
                name: name.trim(),
            },
            spec: {
                profile,
                reportStorage: {
                    configMap: {
                        enabled: true,
                        format: formats.join(','),
                    },
                },
            },
        };

        try {
            await k8sCreate({ model: clusterAssessmentModel, data: resource });
            setIsSubmitting(false);
            setName('');
            setProfile('production');
            onCreated();
            onClose();
        } catch (err) {
            setIsSubmitting(false);
            setError(err instanceof Error ? err.message : String(err));
        }
    };

    const handleClose = () => {
        setName('');
        setProfile('production');
        setError(null);
        onClose();
    };

    return (
        <Modal
            variant={ModalVariant.medium}
            title="Create Cluster Assessment"
            isOpen={isOpen}
            onClose={handleClose}
            actions={[
                <Button
                    key="create"
                    variant="primary"
                    onClick={handleSubmit}
                    isDisabled={isSubmitting || !name.trim() || (!enableHtml && !enableJson && !enablePdf)}
                    isLoading={isSubmitting}
                >
                    Create
                </Button>,
                <Button
                    key="cancel"
                    variant="link"
                    onClick={handleClose}
                    isDisabled={isSubmitting}
                >
                    Cancel
                </Button>,
            ]}
        >
            <Form>
                {error && (
                    <Alert variant="danger" isInline title="Error creating assessment">
                        {error}
                    </Alert>
                )}
                <FormGroup label="Name" isRequired fieldId="assessment-name">
                    <TextInput
                        isRequired
                        autoFocus
                        id="assessment-name"
                        value={name}
                        onChange={(_event, value) => setName(value)}
                        placeholder="my-assessment"
                        isDisabled={isSubmitting}
                    />
                </FormGroup>
                <FormGroup label="Profile" fieldId="assessment-profile">
                    <FormSelect
                        id="assessment-profile"
                        value={profile}
                        onChange={(_event, value) => setProfile(value)}
                        isDisabled={isSubmitting}
                    >
                        {profileOptions.map((option) => (
                            <FormSelectOption
                                key={option.value}
                                value={option.value}
                                label={option.label}
                            />
                        ))}
                    </FormSelect>
                    <FormHelperText>
                        <HelperText>
                            <HelperTextItem variant="default">
                                {selectedProfileDescription}
                            </HelperTextItem>
                        </HelperText>
                    </FormHelperText>
                </FormGroup>
                <FormGroup label="Report Formats" role="group">
                    <Checkbox
                        id="format-html"
                        label="HTML Report"
                        isChecked={enableHtml}
                        onChange={(_event, checked) => setEnableHtml(checked)}
                        isDisabled={isSubmitting}
                    />
                    <Checkbox
                        id="format-json"
                        label="JSON Report"
                        isChecked={enableJson}
                        onChange={(_event, checked) => setEnableJson(checked)}
                        isDisabled={isSubmitting}
                    />
                    <Checkbox
                        id="format-pdf"
                        label="PDF Report"
                        isChecked={enablePdf}
                        onChange={(_event, checked) => setEnablePdf(checked)}
                        isDisabled={isSubmitting}
                    />
                    {(!enableHtml && !enableJson && !enablePdf) && (
                        <FormHelperText>
                            <HelperText>
                                <HelperTextItem variant="error" icon={<ExclamationCircleIcon />}>
                                    At least one report format must be selected.
                                </HelperTextItem>
                            </HelperText>
                        </FormHelperText>
                    )}
                </FormGroup>
            </Form>
        </Modal>
    );
}
