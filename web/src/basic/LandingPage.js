import React from "react";
import * as Setting from "../Setting";
import {Helmet} from "react-helmet";
import "./LandingPage.css";

const COMPANY_SHORT_NAME = "RH";
const COMPANY_FULL_NAME = "Neural Architecture";
const HERO_TITLE = "The Neural Architecture of Tomorrow's Intelligence";
const HERO_DESCRIPTION =
  "A young team of experts specializing in AI tools, building a unified ecosystem for next-generation agents.";
const EXTERNAL_WEBSITE_LINK = "https://example.com";
const LOGIN_LINK = "/login/built-in";
const COPYRIGHT_TEXT = `© 2024 ${COMPANY_SHORT_NAME} • ${COMPANY_FULL_NAME}`;

const LandingPage = () => {
  const goToLogin = () => {
    Setting.goToLink(LOGIN_LINK);
  };

  const goToSite = () => {
    window.open(EXTERNAL_WEBSITE_LINK, "_blank");
  };

  return (
    <div
      className="text-on-surface font-body-md min-h-screen"
      style={{
        backgroundColor: "#fafcfd",
        fontFamily:
          "Inter, -apple-system, BlinkMacSystemFont, \"Segoe UI\", Roboto, \"Helvetica Neue\", sans-serif",
        position: "relative",
        overflow: "hidden",
      }}
    >
      <div className="bg-shape1"></div>
      <div className="bg-shape2"></div>
      <Helmet>
        <title>{COMPANY_SHORT_NAME}</title>
      </Helmet>

      <header className="header">
        <div className="header-logo">{COMPANY_SHORT_NAME}</div>
        <button onClick={goToLogin} className="stripe-btn">
          Log in
        </button>
      </header>

      <main>
        <section className="min-h-screen flex flex-col items-center justify-center relative px-margin-mobile overflow-hidden">
          <div className="absolute inset-0 soft-glow pointer-events-none opacity-50"></div>
          <div className="max-w-container-max mx-auto text-center relative z-10 animate-fade-in">
            <div className="mb-12 inline-block">
              <span className="text-primary font-display-xl text-headline-lg tracking-tighter border-b-4 border-primary/20 pb-2">
                {COMPANY_SHORT_NAME}
              </span>
            </div>
            <h1
              className="font-display-xl text-4xl md:text-display-xl text-on-surface mb-8 max-w-4xl mx-auto leading-tight"
              style={{
                fontSize: "64px",
                lineHeight: "1.1",
                letterSpacing: "-0.02em",
                fontWeight: "700",
                fontFamily: "Space Grotesk",
              }}
            >
              {HERO_TITLE}
            </h1>
            <p
              className="font-body-lg text-body-lg text-on-surface-variant mb-12 max-w-2xl mx-auto opacity-80"
              style={{
                fontSize: "18px",
                lineHeight: "1.7",
                fontWeight: "400",
                fontFamily: "Inter",
              }}
            >
              {HERO_DESCRIPTION}
            </p>
            <div className="hero-buttons">
              <button
                onClick={goToSite}
                className="bg-primary text-on-primary px-10 py-4 rounded-full font-label-caps text-label-caps hover:shadow-xl hover:-translate-y-0.5 transition-all active:scale-95"
                style={{
                  backgroundColor: "#006877",
                  color: "#ffffff",
                  fontSize: "12px",
                  lineHeight: "1",
                  letterSpacing: "0.15em",
                  fontWeight: "600",
                  fontFamily: "Space Grotesk",
                  border: "none",
                  cursor: "pointer",
                  padding: "16px 40px",
                  borderRadius: "9999px",
                }}
              >
                Join Ecosystem
              </button>
              <button
                onClick={goToSite}
                className="flex-center"
                style={{
                  color: "#006877",
                  fontSize: "12px",
                  lineHeight: "1",
                  letterSpacing: "0.15em",
                  fontWeight: "600",
                  fontFamily: "Space Grotesk",
                  background: "none",
                  border: "none",
                  cursor: "pointer",
                  padding: "10px",
                }}
              >
                Learn More{" "}
                <span
                  className="material-symbols-outlined"
                  style={{fontSize: "18px"}}
                >
                  arrow_forward
                </span>
              </button>
            </div>
          </div>
        </section>

        <section
          className="py-section-gap px-margin-mobile"
          style={{paddingBottom: "160px", paddingTop: "160px"}}
        >
          <div className="max-w-container-max mx-auto">
            <div className="text-center mb-24 animate-fade-in delay-100">
              <span
                className="text-primary font-label-caps text-label-caps tracking-widest uppercase mb-4 block"
                style={{
                  color: "#006877",
                  fontSize: "12px",
                  letterSpacing: "0.15em",
                  fontWeight: "600",
                  fontFamily: "Space Grotesk",
                }}
              >
                Our Modules
              </span>
              <h2
                className="font-headline-lg text-headline-lg text-on-surface"
                style={{
                  fontSize: "48px",
                  lineHeight: "1.2",
                  letterSpacing: "-0.01em",
                  fontWeight: "600",
                  fontFamily: "Space Grotesk",
                }}
              >
                The Intelligence Foundation
              </h2>
            </div>
            <div className="modules-grid">
              <div className="glass-card p-10 rounded-xl hover:shadow-2xl hover:shadow-primary/5 transition-all animate-fade-in delay-200">
                <div
                  className="w-14 h-14 rounded-full bg-primary-container/30 flex items-center justify-center mb-8"
                  style={{backgroundColor: "rgba(165, 238, 255, 0.3)"}}
                >
                  <span
                    className="material-symbols-outlined text-primary"
                    style={{color: "#006877"}}
                  >
                    hub
                  </span>
                </div>
                <h3
                  className="font-headline-md text-2xl text-on-surface mb-4"
                  style={{
                    fontSize: "32px",
                    lineHeight: "1.3",
                    fontWeight: "500",
                    fontFamily: "Space Grotesk",
                  }}
                >
                  MCP
                </h3>
                <p
                  className="text-on-surface-variant font-body-md opacity-70"
                  style={{
                    fontSize: "16px",
                    lineHeight: "1.7",
                    fontWeight: "400",
                    fontFamily: "Inter",
                  }}
                >
                  Model Context Protocol for standardized data exchange between
                  models and external sources.
                </p>
              </div>
              <div className="glass-card p-10 rounded-xl hover:shadow-2xl hover:shadow-primary/5 transition-all animate-fade-in delay-200">
                <div
                  className="w-14 h-14 rounded-full bg-primary-container/30 flex items-center justify-center mb-8"
                  style={{backgroundColor: "rgba(165, 238, 255, 0.3)"}}
                >
                  <span
                    className="material-symbols-outlined text-primary"
                    style={{color: "#006877"}}
                  >
                    analytics
                  </span>
                </div>
                <h3
                  className="font-headline-md text-2xl text-on-surface mb-4"
                  style={{
                    fontSize: "32px",
                    lineHeight: "1.3",
                    fontWeight: "500",
                    fontFamily: "Space Grotesk",
                  }}
                >
                  Advanced Analytics
                </h3>
                <p
                  className="text-on-surface-variant font-body-md opacity-70"
                  style={{
                    fontSize: "16px",
                    lineHeight: "1.7",
                    fontWeight: "400",
                    fontFamily: "Inter",
                  }}
                >
                  Deep analysis of neural connections and agent performance in
                  real-time.
                </p>
              </div>
              <div className="glass-card p-10 rounded-xl hover:shadow-2xl hover:shadow-primary/5 transition-all animate-fade-in delay-300">
                <div
                  className="w-14 h-14 rounded-full bg-primary-container/30 flex items-center justify-center mb-8"
                  style={{backgroundColor: "rgba(165, 238, 255, 0.3)"}}
                >
                  <span
                    className="material-symbols-outlined text-primary"
                    style={{color: "#006877"}}
                  >
                    memory
                  </span>
                </div>
                <h3
                  className="font-headline-md text-2xl text-on-surface mb-4"
                  style={{
                    fontSize: "32px",
                    lineHeight: "1.3",
                    fontWeight: "500",
                    fontFamily: "Space Grotesk",
                  }}
                >
                  Neural Memory
                </h3>
                <p
                  className="text-on-surface-variant font-body-md opacity-70"
                  style={{
                    fontSize: "16px",
                    lineHeight: "1.7",
                    fontWeight: "400",
                    fontFamily: "Inter",
                  }}
                >
                  Long-term context storage mimicking synaptic connections for
                  continuous learning.
                </p>
              </div>
              <div className="glass-card p-10 rounded-xl hover:shadow-2xl hover:shadow-primary/5 transition-all animate-fade-in delay-300">
                <div
                  className="w-14 h-14 rounded-full bg-primary-container/30 flex items-center justify-center mb-8"
                  style={{backgroundColor: "rgba(165, 238, 255, 0.3)"}}
                >
                  <span
                    className="material-symbols-outlined text-primary"
                    style={{color: "#006877"}}
                  >
                    precision_manufacturing
                  </span>
                </div>
                <h3
                  className="font-headline-md text-2xl text-on-surface mb-4"
                  style={{
                    fontSize: "32px",
                    lineHeight: "1.3",
                    fontWeight: "500",
                    fontFamily: "Space Grotesk",
                  }}
                >
                  Orchestration
                </h3>
                <p
                  className="text-on-surface-variant font-body-md opacity-70"
                  style={{
                    fontSize: "16px",
                    lineHeight: "1.7",
                    fontWeight: "400",
                    fontFamily: "Inter",
                  }}
                >
                  Intelligent management of agent swarms for solving complex
                  multi-level tasks.
                </p>
              </div>
            </div>
          </div>
        </section>

        <section
          className="py-section-gap px-margin-mobile bg-white/50"
          style={{paddingBottom: "160px", paddingTop: "160px"}}
        >
          <div className="max-w-3xl mx-auto text-center animate-fade-in">
            <h2
              className="font-headline-lg text-headline-lg text-on-surface mb-8"
              style={{
                fontSize: "48px",
                lineHeight: "1.2",
                letterSpacing: "-0.01em",
                fontWeight: "600",
                fontFamily: "Space Grotesk",
              }}
            >
              Ready to evolve?
            </h2>
            <p
              className="text-body-lg text-on-surface-variant mb-12 opacity-80"
              style={{
                fontSize: "18px",
                lineHeight: "1.7",
                fontWeight: "400",
                fontFamily: "Inter",
              }}
            >
              Join the first truly integrated AI ecosystem. Early alpha access
              is now available for partners.
            </p>
            <div className="hero-buttons">
              <button
                onClick={goToSite}
                className="bg-on-surface text-surface px-12 py-5 rounded-full font-label-caps text-label-caps hover:bg-primary transition-colors"
                style={{
                  backgroundColor: "#1a1c1e",
                  color: "#ffffff",
                  fontSize: "12px",
                  lineHeight: "1",
                  letterSpacing: "0.15em",
                  fontWeight: "600",
                  fontFamily: "Space Grotesk",
                  border: "none",
                  cursor: "pointer",
                  padding: "20px 48px",
                  borderRadius: "9999px",
                }}
              >
                Request Access
              </button>
              <p
                className="text-on-surface-variant/40 font-label-caps text-[10px] tracking-widest uppercase"
                style={{
                  fontSize: "10px",
                  letterSpacing: "0.15em",
                  fontWeight: "600",
                  fontFamily: "Space Grotesk",
                  opacity: 0.4,
                }}
              >
                {COPYRIGHT_TEXT}
              </p>
            </div>
          </div>
        </section>
      </main>
    </div>
  );
};

export default LandingPage;
