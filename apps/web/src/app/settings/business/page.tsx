import { SettingsSectionPreview } from '@/components/settings-section-preview';
import { findSettingsSection } from '@/fixtures/preview';

const section = findSettingsSection('/settings/business');

export default function BusinessSettingsPage() {
  if (!section) {
    throw new Error('Missing business settings preview fixture.');
  }

  return <SettingsSectionPreview section={section} />;
}
