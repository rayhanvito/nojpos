import { SettingsSectionPreview } from '@/components/settings-section-preview';
import { findSettingsSection } from '@/fixtures/preview';

const section = findSettingsSection('/settings/receipt');

export default function ReceiptSettingsPage() {
  if (!section) {
    throw new Error('Missing receipt settings preview fixture.');
  }

  return <SettingsSectionPreview section={section} />;
}
