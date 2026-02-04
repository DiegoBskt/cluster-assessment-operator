import * as React from 'react';
import {
    Alert,
    AlertVariant,
    Flex,
    FlexItem,
    Label,
} from '@patternfly/react-core';
import { DeltaSummary } from '../types';

interface DeltaBannerProps {
    delta?: DeltaSummary;
}

export default function DeltaBanner({ delta }: DeltaBannerProps) {
    if (!delta) {
        return null;
    }

    const newCount = delta.newFindings?.length ?? 0;
    const resolvedCount = delta.resolvedFindings?.length ?? 0;
    const regressionCount = delta.regressionFindings?.length ?? 0;
    const improvedCount = delta.improvedFindings?.length ?? 0;
    const scoreDelta = delta.scoreDelta;

    // Don't show banner if nothing changed
    if (newCount === 0 && resolvedCount === 0 && regressionCount === 0 && improvedCount === 0 && (scoreDelta === undefined || scoreDelta === 0)) {
        return (
            <Alert variant={AlertVariant.info} isInline isPlain title="No changes from previous run" />
        );
    }

    const variant =
        regressionCount > 0 || newCount > resolvedCount
            ? AlertVariant.warning
            : resolvedCount > 0 || improvedCount > 0
            ? AlertVariant.success
            : AlertVariant.info;

    const scorePart =
        scoreDelta !== undefined && scoreDelta !== 0
            ? `Score: ${scoreDelta > 0 ? '+' : ''}${scoreDelta}`
            : '';

    return (
        <Alert variant={variant} isInline title="Changes from previous assessment" style={{ marginBottom: '16px' }}>
            <Flex>
                {newCount > 0 && (
                    <FlexItem>
                        <Label color="red">{newCount} new finding{newCount !== 1 ? 's' : ''}</Label>
                    </FlexItem>
                )}
                {resolvedCount > 0 && (
                    <FlexItem>
                        <Label color="green">{resolvedCount} resolved</Label>
                    </FlexItem>
                )}
                {regressionCount > 0 && (
                    <FlexItem>
                        <Label color="orange">{regressionCount} regressed</Label>
                    </FlexItem>
                )}
                {improvedCount > 0 && (
                    <FlexItem>
                        <Label color="blue">{improvedCount} improved</Label>
                    </FlexItem>
                )}
                {scorePart && (
                    <FlexItem>
                        <Label color={scoreDelta! > 0 ? 'green' : 'red'}>{scorePart}</Label>
                    </FlexItem>
                )}
            </Flex>
        </Alert>
    );
}
