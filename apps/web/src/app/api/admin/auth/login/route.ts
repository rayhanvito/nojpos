import { NextRequest } from 'next/server';
import { handleAdminLogin } from '../../../../../lib/server/admin-session';

export async function POST(request: NextRequest) {
  return handleAdminLogin(request);
}
