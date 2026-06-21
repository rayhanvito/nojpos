export type PreviewStateTone = 'loading' | 'empty' | 'error' | 'forbidden' | 'unavailable';

export type PreviewLane = {
  label: string;
  items: readonly string[];
};

export type PreviewStateFixture = {
  tone: PreviewStateTone;
  title: string;
  message: string;
};

export type PreviewMetricFixture = {
  label: string;
  value: string;
  description: string;
  badge?: string;
  tone?: 'neutral' | 'success' | 'warning';
};

export type PreviewTableFixture = {
  columns: readonly string[];
  rows: readonly (readonly string[])[];
  caption?: string;
  primaryColumn?: string | number;
  statusColumn?: string | number;
  metaColumns?: readonly (string | number)[];
  numericColumns?: readonly (string | number)[];
};

export type ReadonlyResourceFixture = {
  label: string;
  path: string;
};

export type DisabledFieldFixture = {
  label: string;
  value: string;
  helper: string;
};

export type SettingsSectionFixture = {
  title: string;
  kicker: string;
  badge: string;
  description: string;
  path: string;
  fields: readonly DisabledFieldFixture[];
};

export type DetailActionFixture = {
  label: string;
  status: string;
};

export type PreviewActionCategory = 'read' | 'write' | 'import' | 'export' | 'danger';

export type PreviewActionFixture = {
  label: string;
  category: PreviewActionCategory;
  reason: string;
};

export type PreviewActionGroupFixture = {
  title: string;
  description: string;
  actions: readonly PreviewActionFixture[];
};

export type DetailSectionFixture = {
  title: string;
  rows: readonly DisabledFieldFixture[];
};

export type DetailPreviewFixture = {
  id: string;
  title: string;
  kicker: string;
  badge: string;
  description: string;
  backHref: string;
  backLabel: string;
  actions: readonly DetailActionFixture[];
  sections: readonly DetailSectionFixture[];
  timeline: readonly string[];
};
