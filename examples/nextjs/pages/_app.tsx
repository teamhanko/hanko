import "../styles/index.css";
import Head from 'next/head';
import type { AppProps } from "next/app";

function MyApp({ Component, pageProps }: AppProps) {
  return <>
    <Head>
      <title>Hanko Next.js Example</title>
      <link rel="icon" href="favicon.png" />
    </Head>
    <Component {...pageProps} />
  </>
}

export default MyApp;
