import { NextRequest } from 'next/server';

import { handleAdminTransactions } from '../../../../lib/server/transactions';

export async function GET(request: NextRequest) {
  return handleAdminTransactions(request);
}
