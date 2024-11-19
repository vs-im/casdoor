// types.js

/**
 * @typedef {Object} Provider
 * @property {string} owner
 * @property {string} name
 * @property {boolean} canSignUp
 * @property {boolean} canSignIn
 * @property {boolean} canUnlink
 * @property {(string|null)} countryCodes
 * @property {boolean} prompted
 * @property {string} signupGroup
 * @property {string} rule
 * @property {Object} provider
 * @property {string} provider.owner
 * @property {string} provider.name
 * @property {string} provider.createdTime
 * @property {string} provider.displayName
 * @property {string} provider.category
 * @property {string} provider.type
 * @property {string} provider.subType
 * @property {string} provider.method
 * @property {string} provider.clientId
 * @property {string} provider.clientSecret
 * @property {string} provider.clientId2
 * @property {string} provider.clientSecret2
 * @property {string} provider.cert
 * @property {string} provider.customAuthUrl
 * @property {string} provider.customTokenUrl
 * @property {string} provider.customUserInfoUrl
 * @property {string} provider.customLogo
 * @property {string} provider.scopes
 * @property {(Object|null)} provider.userMapping
 * @property {string} provider.host
 * @property {number} provider.port
 * @property {boolean} provider.disableSsl
 * @property {string} provider.title
 * @property {string} provider.content
 * @property {string} provider.receiver
 * @property {string} provider.regionId
 * @property {string} provider.signName
 * @property {string} provider.templateCode
 * @property {string} provider.appId
 * @property {string} provider.endpoint
 * @property {string} provider.intranetEndpoint
 * @property {string} provider.domain
 * @property {string} provider.bucket
 * @property {string} provider.pathPrefix
 * @property {string} provider.metadata
 * @property {string} provider.idP
 * @property {string} provider.issuerUrl
 * @property {boolean} provider.enableSignAuthnRequest
 * @property {string} provider.providerUrl
 */

/**
 * @typedef {Object} SignupItem
 * @property {string} name
 * @property {boolean} visible
 * @property {boolean} required
 * @property {boolean} prompted
 * @property {string} type
 * @property {string} customCss
 * @property {string} label
 * @property {string} placeholder
 * @property {(Array|null)} options
 * @property {string} regex
 * @property {string} rule
 */

/**
 * @typedef {Object} SigninMethod
 * @property {string} name
 * @property {string} displayName
 * @property {string} rule
 */

/**
 * @typedef {Object} Organization
 * @property {string} owner
 * @property {string} name
 * @property {string} createdTime
 * @property {string} displayName
 * @property {string} websiteUrl
 * @property {string} logo
 * @property {string} logoDark
 * @property {string} favicon
 * @property {string} passwordType
 * @property {string} passwordSalt
 * @property {Array<string>} passwordOptions
 * @property {string} passwordObfuscatorType
 * @property {string} passwordObfuscatorKey
 * @property {Array<string>} countryCodes
 * @property {string} defaultAvatar
 * @property {string} defaultApplication
 * @property {Array<string>} tags
 * @property {Array<string>} languages
 * @property {null} themeData
 * @property {string} masterPassword
 * @property {string} defaultPassword
 * @property {string} masterVerificationCode
 * @property {string} ipWhitelist
 * @property {number} initScore
 * @property {boolean} enableSoftDeletion
 * @property {boolean} isProfilePublic
 * @property {boolean} useEmailAsUsername
 * @property {boolean} enableTour
 * @property {string} ipRestriction
 * @property {Array<Object>} accountItems
 * @property {string} accountItems.name
 * @property {boolean} accountItems.visible
 * @property {string} accountItems.viewRule
 * @property {string} accountItems.modifyRule
 * @property {string} accountItems.regex
 */

/**
 * @typedef {Object} ApplicationResponse
 * @property {ApplicationConfig} data - The application configuration.
 */

/**
 * @typedef {Object} ApplicationConfig
 * @property {string} owner
 * @property {string} name
 * @property {string} createdTime
 * @property {string} displayName
 * @property {string} logo
 * @property {string} homepageUrl
 * @property {string} description
 * @property {string} organization
 * @property {string} cert
 * @property {string} headerHtml
 * @property {boolean} enablePassword
 * @property {boolean} enableSignUp
 * @property {boolean} enableSigninSession
 * @property {boolean} enableAutoSignin
 * @property {boolean} enableCodeSignin
 * @property {boolean} enableSamlCompress
 * @property {boolean} enableSamlC14n10
 * @property {boolean} enableSamlPostBinding
 * @property {boolean} useEmailAsSamlNameId
 * @property {boolean} enableWebAuthn
 * @property {boolean} enableLinkWithEmail
 * @property {string} orgChoiceMode
 * @property {string} samlReplyUrl
 * @property {Array<Provider>} providers
 * @property {Array<SigninMethod>} signinMethods
 * @property {Array<SignupItem>} signupItems
 * @property {Array<SignupItem>} signinItems
 * @property {Array<string>} grantTypes
 * @property {Organization} organizationObj
 * @property {string} certPublicKey
 * @property {Array<string>} tags
 * @property {null} samlAttributes
 * @property {boolean} isShared
 * @property {string} ipRestriction
 * @property {string} clientId
 * @property {string} clientSecret
 * @property {Array<string>} redirectUris
 * @property {string} tokenFormat
 * @property {string} tokenSigningMethod
 * @property {Array<string>} tokenFields
 * @property {number} expireInHours
 * @property {number} refreshExpireInHours
 * @property {string} signupUrl
 * @property {string} signinUrl
 * @property {string} forgetUrl
 * @property {string} affiliationUrl
 * @property {string} ipWhitelist
 * @property {string} termsOfUse
 * @property {string} signupHtml
 * @property {string} signinHtml
 * @property {null} themeData
 * @property {string} footerHtml
 * @property {string} formCss
 * @property {string} formCssMobile
 * @property {number} formOffset
 * @property {string} formSideHtml
 * @property {string} formBackgroundUrl
 * @property {number} failedSigninLimit
 * @property {number} failedSigninFrozenTime
 */
