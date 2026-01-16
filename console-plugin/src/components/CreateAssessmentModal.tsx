import * as React from 'react';
import {
    Modal,
    ModalVariant,
    Button,
    Form,
    FormGroup,
    TextInput,
    Select,
    SelectOption,
    SelectVariant,
    Checkbox,
    ActionGroup,
    Alert,
} from '@patternfly/react-core';
import { k8sCreate } from '@openshift-console/dynamic-plugin-sdk';

interface CreateAssessmentModalProps {
    isOpen: boolean;
    onClose: () => void;
    onCreated: () => void;
}

const clusterAssessmentModel = {
    apiVersion: 'assessment.openshift.io/v1alpha1',
    apiGroup: 'assessment.openshift.io',
    kind: 'ClusterAssessment',
    plural: 'clusterassessments',
};

const CreateAssessmentModal: React.FC<CreateAssessmentModalProps> = ({
    isOpen,
    onClose,
    onCreated,
}) => {
    const [name, setName] = React.useState('');
    const [profile, setProfile] = React.useState('production');
    const [profileOpen, setProfileOpen] = React.useState(false);
    const [enableHtml, setEnableHtml] = React.useState(true);
    const [enableJson, setEnableJson] = React.useState(true);
    const [isSubmitting, setIsSubmitting] = React.useState(false);
    const [error, setError] = React.useState<string | null>(null);

    const handleSubmit = async () => {
        if (!name.trim()) {
            setError('Name is required');
            return;
        }

        setIsSubmitting(true);
        setError(null);

        const formats: string[] = [];
        if (enableHtml) formats.push('html');
        if (enableJson) formats.push('json');

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
                    isDisabled={isSubmitting || !name.trim()}
                    isLoading={isSubmitting}
                >
                    Create
                </Button>,
                <Button key="cancel" variant="link" onClick={handleClose}>
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
                        id="assessment-name"
                        value={name}
                        onChange={(_event, value) => setName(value)}
                        placeholder="my-assessment"
                    />
                </FormGroup>
                <FormGroup label="Profile" fieldId="assessment-profile">
                    <Select
                        id="assessment-profile"
                        variant={SelectVariant.single}
                        isOpen={profileOpen}
                        onToggle={(_event, isExpanded) => setProfileOpen(isExpanded)}
                        onSelect={(_event, selection) => {
                            setProfile(selection as string);
                            setProfileOpen(false);
                        }}
                        selections={profile}
                    >
                        <SelectOption value="production">
                            Production (Strict)
                        </SelectOption>
                        <SelectOption value="development">
                            Development (Relaxed)
                        </SelectOption>
                    </Select>
                </FormGroup>
                <FormGroup label="Report Formats" fieldId="report-formats">
                    <Checkbox
                        id="format-html"
                        label="HTML Report"
                        isChecked={enableHtml}
                        onChange={(_event, checked) => setEnableHtml(checked)}
                    />
                    <Checkbox
                        id="format-json"
                        label="JSON Report"
                        isChecked={enableJson}
                        onChange={(_event, checked) => setEnableJson(checked)}
                    />
                </FormGroup>
            </Form>
        </Modal>
    );
};

export default CreateAssessmentModal;
