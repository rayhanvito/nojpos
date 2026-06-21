import { NextRequest } from 'next/server';

import { handleAdminStaff } from '../../../../lib/server/staff';

export function GET(request: NextRequest) {
  return handleAdminStaff(request);
}
