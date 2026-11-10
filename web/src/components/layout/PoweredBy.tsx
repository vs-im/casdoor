import * as Conf from "@/Conf";

/**
 * The default footer line. Upstream showed "Powered by" and its own wordmark
 * here; this deployment shows the product name only, no outbound links. A
 * deployment can still replace the whole line through `Conf.CustomFooter` or an
 * organization's `footerHtml`. `dark` is for a footer on a dark backdrop while
 * the page itself is light.
 */
export function PoweredBy({dark = false}: {dark?: boolean}) {
  if (Conf.CustomFooter !== null) {
    return <>{Conf.CustomFooter}</>;
  }

  return <span className={dark ? "text-white/80" : undefined}>© {new Date().getFullYear()} {Conf.BrandName}</span>;
}
