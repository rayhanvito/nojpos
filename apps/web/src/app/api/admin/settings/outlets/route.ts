import { NextRequest } from 'next/server';

import { handleAdminSettings } from '../../../../../lib/server/settings-readonly';

export function GET(request: NextRequest) {
  return handleAdminSettings(request, 'outlets');
}
