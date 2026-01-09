import React from 'react';
import { Controller, useForm } from 'react-hook-form';

import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod/v4';

import { InfoCircledIcon } from '@radix-ui/react-icons';
import { Box, Button, Callout, Card, Code, Flex, Select, Text, TextField } from '@radix-ui/themes';

const CITY_OPTIONS = [
  { value: 'Sydney', label: 'Sydney', flag: '🇦🇺' },
  { value: 'Melbourne', label: 'Melbourne', flag: '🇦🇺' },
  { value: 'Brisbane', label: 'Brisbane', flag: '🇦🇺' },
  { value: 'Perth', label: 'Perth', flag: '🇦🇺' },
  { value: 'Adelaide', label: 'Adelaide', flag: '🇦🇺' },
];

const workflowFormSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, { message: 'Name is required' })
    .max(50, { message: 'Name must be less than 50 characters' }),
  email: z.email('Invalid email address'),
  city: z
    .string()
    .trim()
    .min(1, { message: 'City is required' })
    .max(100, { message: 'City must be less than 100 characters' }),
  operator: z.enum([
    'greater_than',
    'less_than',
    'equals',
    'greater_than_or_equal',
    'less_than_or_equal',
  ]),
  threshold: z
    .number()
    .gt(-100, { message: 'Temperature must be above -100°C' })
    .lt(100, { message: 'Temperature must be below 100°C' }),
});

type WorkflowFormData = z.infer<typeof workflowFormSchema>;

interface UserInputFormProps {
  onExecute: (formData: WorkflowFormData) => Promise<void>;
  isExecuting: boolean;
}

export const UserInputForm: React.FC<UserInputFormProps> = ({ onExecute, isExecuting }) => {
  const {
    register,
    handleSubmit,
    formState: { errors },
    control,
    watch,
  } = useForm<WorkflowFormData>({
    resolver: zodResolver(workflowFormSchema),
    defaultValues: {
      name: '',
      email: '',
      city: 'Sydney',
      operator: 'greater_than',
      threshold: 25,
    },
    mode: 'onBlur',
    reValidateMode: 'onChange',
  });

  const watchedValues = watch();

  const operatorLabels = {
    greater_than: 'is greater than',
    less_than: 'is less than',
    equals: 'equals exactly',
    greater_than_or_equal: 'is at least',
    less_than_or_equal: 'is at most',
  };

  const getOperatorSymbol = (operator: keyof typeof operatorLabels) => {
    const symbols = {
      greater_than: '>',
      less_than: '<',
      equals: '=',
      greater_than_or_equal: '≥',
      less_than_or_equal: '≤',
    } as const;
    return symbols[operator] || '>';
  };

  const onSubmit = async (data: WorkflowFormData) => {
    await onExecute(data);
  };

  const handleFormSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    void handleSubmit(onSubmit)(e);
  };

  return (
    <Box p="4" style={{ height: '100%', overflowY: 'auto' }}>
      <Text size="5" weight="bold" mb="4" style={{ display: 'block' }}>
        📝 Configure Weather Alert
      </Text>

      <form onSubmit={handleFormSubmit}>
        <Card mb="4">
          <Box p="4">
            <Text size="4" weight="medium" mb="3" style={{ display: 'block' }}>
              👤 Your Details
            </Text>

            <Flex direction="column" gap="3">
              <Box>
                <Text size="2" weight="medium" mb="1" style={{ display: 'block' }}>
                  Name *
                </Text>
                <TextField.Root {...register('name')} placeholder="Your name" />
                {errors.name && (
                  <Text size="1" color="red" mt="1" style={{ display: 'block' }}>
                    {errors.name.message}
                  </Text>
                )}
              </Box>

              <Box>
                <Text size="2" weight="medium" mb="1" style={{ display: 'block' }}>
                  Email *
                </Text>
                <TextField.Root
                  {...register('email')}
                  type="email"
                  placeholder="your.email@example.com"
                />
                {errors.email && (
                  <Text size="1" color="red" mt="1" style={{ display: 'block' }}>
                    {errors.email.message}
                  </Text>
                )}
              </Box>

              <Box>
                <Text size="2" weight="medium" mb="1" style={{ display: 'block' }}>
                  City *
                </Text>
                <Controller
                  name="city"
                  control={control}
                  render={({ field }) => (
                    <Select.Root value={field.value} onValueChange={field.onChange}>
                      <Select.Trigger />
                      <Select.Content>
                        {CITY_OPTIONS.map(city => (
                          <Select.Item key={city.value} value={city.value}>
                            {city.label}
                          </Select.Item>
                        ))}
                      </Select.Content>
                    </Select.Root>
                  )}
                />
                {errors.city && (
                  <Text size="1" color="red" mt="1" style={{ display: 'block' }}>
                    {errors.city.message}
                  </Text>
                )}
              </Box>
            </Flex>
          </Box>
        </Card>

        <Card mb="4">
          <Box p="4">
            <Text size="4" weight="medium" mb="3" style={{ display: 'block' }}>
              🌡️ Alert Condition
            </Text>

            <Text size="2" color="gray" mb="3" style={{ display: 'block' }}>
              When should we send you an alert?
            </Text>

            <Flex direction="column" gap="3">
              <Flex align="center" gap="2" wrap="wrap">
                <Text size="2">Send alert when temperature</Text>
                <Controller
                  name="operator"
                  control={control}
                  render={({ field }) => (
                    <Select.Root value={field.value} onValueChange={field.onChange}>
                      <Select.Trigger style={{ minWidth: '140px' }} />
                      <Select.Content>
                        {Object.entries(operatorLabels).map(([value, label]) => (
                          <Select.Item key={value} value={value}>
                            {label}
                          </Select.Item>
                        ))}
                      </Select.Content>
                    </Select.Root>
                  )}
                />

                <TextField.Root
                  {...register('threshold', { valueAsNumber: true })}
                  type="number"
                  style={{ width: '80px' }}
                />
                <Text size="2">°C</Text>
              </Flex>

              {errors.operator && (
                <Text size="1" color="red" style={{ display: 'block' }}>
                  {errors.operator.message}
                </Text>
              )}
              {errors.threshold && (
                <Text size="1" color="red" style={{ display: 'block' }}>
                  {errors.threshold.message}
                </Text>
              )}
            </Flex>

            <Callout.Root color="blue" variant="soft" mt="3">
              <Callout.Icon>
                <InfoCircledIcon />
              </Callout.Icon>
              <Callout.Text>
                <Text size="2" weight="medium">
                  Condition:{' '}
                </Text>
                <Code size="2">
                  temperature {getOperatorSymbol(watchedValues.operator)} {watchedValues.threshold}
                  °C
                </Code>
              </Callout.Text>
            </Callout.Root>
          </Box>
        </Card>

        <Button
          type="submit"
          size="3"
          style={{ width: '100%' }}
          disabled={isExecuting}
          loading={isExecuting}
        >
          Execute Workflow
        </Button>
      </form>

      <Card mt="4">
        <Box p="3">
          <Text size="2" weight="medium" mb="2" style={{ display: 'block' }}>
            Configuration Preview:
          </Text>
          <Code size="1" style={{ display: 'block', whiteSpace: 'pre-wrap' }}>
            {JSON.stringify(watchedValues, null, 2)}
          </Code>
        </Box>
      </Card>
    </Box>
  );
};
