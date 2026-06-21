import { SettingsSectionPreview } from '@/components/settings-section-preview';
import { findSettingsSection } from '@/fixtures/preview';

const section = findSettingsSection('/settings/outlet');

export default function OutletSettingsPage() {
  if (!section) {
    throw new Error('Missing outlet settings preview fixture.');
  }

  return <SettingsSectionPreview section={section} />;
}
