import { NextRequest } from 'next/server';

import { handleAdminCatalogProducts } from '../../../../../lib/server/catalog';

export async function GET(request: NextRequest) {
  return handleAdminCatalogProducts(request);
}
