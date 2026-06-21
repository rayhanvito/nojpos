import { SettingsSectionPreview } from '@/components/settings-section-preview';
import { findSettingsSection } from '@/fixtures/preview';

const section = findSettingsSection('/settings/payments');

export default function PaymentSettingsPage() {
  if (!section) {
    throw new Error('Missing payment settings preview fixture.');
  }

  return <SettingsSectionPreview section={section} />;
}
