import { NextRequest } from 'next/server';

import { handleAdminCustomers } from '../../../../lib/server/customers';

export function GET(request: NextRequest) {
  return handleAdminCustomers(request);
}
