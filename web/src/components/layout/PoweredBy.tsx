import * as Conf from "@/Conf";

/**
 * The default footer line. Upstream showed "Powered by" and its own wordmark
 * here; this deployment shows the product name only, no outbound links. A
 * deployment can still replace the whole line through `Conf.CustomFooter` or an
 * organization's `footerHtml`.
 */
export function PoweredBy() {
  if (Conf.CustomFooter !== null) {
    return <>{Conf.CustomFooter}</>;
  }

  return <span>© {new Date().getFullYear()} {Conf.ProductName}</span>;
}
