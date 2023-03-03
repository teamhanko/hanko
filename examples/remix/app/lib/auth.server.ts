import { redirect } from '@remix-run/server-runtime';
import { parse } from 'cookie';
import type { JwtPayload } from 'jsonwebtoken';
import { decode } from 'jsonwebtoken';

export const extractHankoCookie = (request: Request) => {
    const cookies = parse(request.headers.get('Cookie') || '');
    return cookies.hanko;
};

// ensures the user has a hanko cookie but does not check if it is valid
export async function requireValidJwt(request: Request) {
    const hankoCookie = extractHankoCookie(request);
    const decoded = decode(hankoCookie) as JwtPayload;
    const hankoId = decoded?.sub;
    const exp = (decoded?.exp || 0) * 1000;
    if (!hankoId || exp < Date.now())
        throw redirect(`/`);
    return decoded
}
