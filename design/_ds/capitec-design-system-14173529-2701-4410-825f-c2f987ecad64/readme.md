# Capitec Design System

The design system for Capitec Bank's onboarding and customer-facing interfaces, extracted from the "Minor Onboarding" Figma file.

## Overview

This design system captures the visual language, components, and interaction patterns used across Capitec's digital products, with a focus on onboarding flows for minor account holders and guardian registration.

## Source Materials

- **Figma file**: Minor Onboarding.fig (mounted as virtual filesystem)
- **Pages included**: Components, Minor-Onboarding-Prototype, Copy-Of-FNO-onboarding-of-UT-prototype---Columbine, User-Research-Testing, and supporting archives
- **Total pages**: 13 pages with 111 frames
- **Component families**: 405 families with 649 component variants
- **Token collections**: 5 collections with 824 variables across multiple modes
- **Typography**: Text and effect styles from Nunito Sans, Hind Vadodara, Inter, and Roboto families

## Visual Foundations

### Colors

Capitec's color system is built on Material Design 3 principles with extensive theming support:

- **Primary blue**: `#2F70EF` (rgb(47, 112, 239)) — the brand's core accent color used throughout interactive elements, links, and primary actions
- **Neutral palette**: From `#FFFFFF` (white) through `#1F272E` (near-black), supporting surfaces, text, and dividers
- **Semantic colors**: Green (#369C3C success), red (#F9423A error), orange (#FFD966 warning)
- **Surfaces**: Tonal backgrounds range from `#F8FAFV` to `#E0E5F2` for layering and depth

The system includes 40+ theme modes: Light/Dark themes, Kendo breakpoints (Regular 0%, Medium 130%, Large 170%, Extra Large 200%), Capitec App variant, IC (Internal Comms) themes, and test modes.

### Typography

A sophisticated multi-family system scaled for accessibility:

- **Nunito Sans** (primary): Bold, SemiBold, Regular, ExtraBold weights across 12–66px sizes
- **Nunito** (secondary): Used for larger display sizes (24–240px) and secondary content
- **Hind Vadodara** (tertiary): Hindi script support for regional variants
- **Inter** (supplementary): Medium and Bold for select UI elements
- **Roboto** (fallback): For compatibility

Common sizes: 12px (smallest UI text), 14px (labels/helper text), 16px (body/form input), 18px (subheadings), 24px (section headers), 32–56px (display sizes), 66–96px (hero text).

### Spacing

Capitec uses an 8px base grid across padding, margins, gaps, and icon sizing. Token values range from 4px to 256px with semantic names (e.g., `--spacing-xs`, `--spacing-lg`).

### Component Visual Language

- **Buttons**: Material Design 3 with Filled, Tonal, Outlined, and Text variants; consistent padding (12px vertical, 16–24px horizontal for standard size)
- **Input fields**: Soft 8px border radius, 12px padding, supporting icon placeholders on leading and trailing edges
- **Cards & surfaces**: 8px rounded corners by default, with drop shadows using `rgba(0, 51, 160, 0.15)` for subtle depth
- **Icons**: 24px default size (also 16px, 18px, 32px variants); single-color glyphs that respond to `currentColor`
- **Dividers**: Thin horizontal/vertical lines at `#CFD5E0` or darker neutrals depending on context
- **States**: Hover (lightened surface), active/pressed (darkened or saturated color), disabled (muted, no interaction)

### Motion & Interaction

- No extraneous animations; interactions use brief, purposeful easing
- State changes are instant or fade-in quickly; longer transitions reserved for purposeful reveals
- Mobile gestures: tap targets minimum 48px, swipe areas clearly defined
- Feedback is immediate and clear (spinners, success states, error callouts)

### Imagery & Iconography

- **Hero images**: Full-bleed photography at tablet/desktop breakpoints (skipped on mobile per space constraints)
- **Icons**: 143-glyph system covering common actions (add, delete, search, settings, etc.), statuses (success, error, pending, info), and product indicators (transactions, credit, insurance, etc.)
- **No brand mark in source**: The Figma file contains no logo. Where a mark is needed, use the brand name rendered in Nunito Sans ExtraBold at appropriate display size
- **Illustrations**: Biometric authentication, hero scenes, and empty states are included as bitmap fills; originals sit in `components/assets/` and `components/screens/assets/`

### Responsive Breakpoints

- **Mobile** (320–599px): Single column, nav drawer, full-width sheets
- **Tablet** (600–839px): 2-column or drawer + content layout
- **Desktop** (840px+): Multi-column with persistent navigation rail

## Content Fundamentals

### Language & Tone

Capitec's copy is:

- **Conversational but clear**: "Set up your account" not "Initialize user profile"
- **Action-oriented**: Verbs lead; "Add member" not "Member addition"
- **Respectful of privacy**: "Securely verify" not "We will spy on you"
- **Accessible**: Plain language, short sentences, defined jargon
- **I/You balance**: First-person in onboarding ("I am a guardian"), second-person for instructions ("You'll need your ID")
- **Emoji**: None in the core UI; reserved for notification badges and promotional contexts only
- **Casing**: Sentence case for headers and labels ("Guardian information" not "Guardian Information"), all caps avoided except acronyms (ID, PIN, NFC)
- **Numbers**: Written out 0–9, digits 10+; currency symbols inline with figures ($100, not $ 100)
- **Punctuation**: No periods on button labels or error callouts; commas in lists; em dashes (—) for sidenotes

## Components

This design system implements **333 of 506 component families** from the Figma kit. The remaining 173 are intentionally skipped as duplicates, test frames, or WIP placeholders. See [SKIPPED_COMPONENTS.md](./SKIPPED_COMPONENTS.md) for details.

### Custom Additions

Three components are intentional additions not in the Figma kit vocabulary. These are **confirmed production-ready**:
- **AgentSignInFingerPrint** — biometric authentication variant for agent onboarding
- **AgentSignInLoading** — loading state for agent sign-in flow
- **SAMinorSAGuardian** — South Africa–specific minor/guardian selection screen

All three support the onboarding and authentication flows and are fully exported to `window.CapitecDesignSystem_141735`.

### Core Families

The 333 built families exported to `window.CapitecDesignSystem_141735`:

AccessFacility, AccountListItem, AccountTile, Add, Add2, Add4, AdjustAnimationUp, AdjustmentArrow, AgentSignInFingerPrint, AgentSignInLoading, AgentSignInSuccess, Alert, AlertDialog, AlertIcon, AllInclusive, AnnualPercentageRate, AppBar, ApplicationMenu, ArrowDown, ArrowLeft, ArrowRight, ArrowUp, AssetTagging, AssignmentCompleted, AssignmentPending, AssistiveListening, AssistiveListeningOff, Audio, AudioDescription, AudioFile, AudioSpeaker, AutoAwesome, AutoDelete, AutoMode, Autorenew, AvTimer, Badge, BakingCookie, Ballot, BannerAdPx, BannerAd, BarChart, Batch, BatteryAlert, BatteryChargingFull, BatteryFull, BatteryLow, BatteryProfileManager, BatteryStatusBad, BatteryStatusGood, BatteryStd, BatteryUnknown, BeachAccess, BeautyCosmeticsIcon, Bed, BedroomBaby, BedroomParent, Beenhere, BeerIcon, BetterThanExpectedOutcome, BgColor, BikeScooter, Biotech, Blender, BlindsClosed, BlindsOpen, BlockFlipped, Block, BloodtypeIcon, Blossom, BluetoothAudio, BluetoothConnected, BluetoothDisabled, BluetoothDrive, BluetoothOff, BluetoothSearching, BlurCircular, BlurLinear, BlurOn, BoardInfinity, Boats, BodyAwareness, BodyScan, Bolt, BoltIcon, Bomb, BoneIcon, BookIcon, Bookmark, BookmarkAdd, BookmarkAdded, BookmarkBorder, BookmarkOutline, BookmarkRemove, BookmarkRemoved, Books, BooksIcon, Bookstore, BorderAll, BorderBottom, BorderClear, BorderColor, BorderHorizontal, BorderInner, BorderLeft, BorderOuter, BorderRight, BorderStyle, BorderTop, BorderVertical, BoundaryIcon, BoxIcon, BrandingWatermark, Brass, BrassIcon, BrainIcon, Brakes, BranchIcon, BrokenImageIcon, BroomIcon, BrushIcon, BubbleChart, BubblesIcon, BugReportIcon, BuildIcon, BuildCircleIcon, Builder, BuilderIcon, BuildingIcon, BulkSelectionCheckbox, BulkSmsIcon, BulletIcon, BullseyeIcon, BuoyIcon, BurgerMenuIcon, BurgerIcon, BuriedIconContainer, BusFareIcon, BusIcon, BusinessCenterIcon, BusinessIcon, BusinessMessagesIcon, BustPortraitIcon, BustinessCardIcon, ButtonIcon, ButtonGroupIcon, BuyIcon, BuysIcon, CabinIcon, CableIcon, CachedIcon, CactusIcon, CadetIcon, CafeIcon, CakeIcon, CalculateIcon, CalculatorIcon, CalendarMonthIcon, CalendarTodayIcon, CalendarViewDayIcon, CalendarViewMonthIcon, CalendarViewWeekIcon, CalendarViewIcon, CameraAltIcon, CameraEnhanceIcon, CameraFrontIcon, CameraIcon, CameraIndoorIcon, CameraOutdoorIcon, CameraRollIcon, CampIcon, CampsiteIcon, CanalIcon, CancelIcon, CancelPresentationIcon, CancelScheduleSendIcon, CancellingIcon, CandlestickChartIcon, CandyIcon, CaneIcon, CanineIcon, CanIsland, CanoeIcon, CanPlantIcon, CanSliderIcon, CannedIcon, CannyIcon, CantIcon, CanteenIcon, CantinaIcon, CantonIcon, CanvasIcon, CanvasesIcon, CanvasSupportIcon, CanvasWithObjectsIcon, CapsuleIcon, CarRentalIcon, CarRepairIcon, CarpenterIcon, CartIcon, CartographyIcon, CaseIcon, CasinoIcon, CasketIcon, CastIcon, CastConnectedIcon, CastForEducationIcon, CastleIcon, CasualDiningIcon, CatalogIcon, CatalogueIcon, CatalogsIcon, CatchingPokemonIcon, CategoricalVariableIcon, CategoryIcon, CathedralsIcon, CaterpillarIcon, CatsIcon, CattleIcon, CausewayIcon, CautionIcon, CaviarIcon, CavityIcon, CdIcon, CdRomIcon, CedarIcon, CeilidhIcon, CeilingIcon, CelIcon, CelebrateIcon, CelebrationIcon, CelebrityIcon, CelebsIcon, CeleritasIcon, CelestialIcon, CelesticsIcon, CeleryIcon, CelestialNavigationIcon, CelestialObjectsIcon, CelestialSphereIcon, CeleusIcon, CelIdaIcon, CeliacDiseaseIcon, CelidoneIcon, CelidorIcon, CelieIcon, CelieiaIcon, CelierIcon, CeliestIcon, CelieusIcon, CelieusaIcon, CelieuxIcon, CeliesIcon, CelieaIcon, CelieceIcon, CeliedIcon, CeliexIcon, CelifexIcon, CelieineIcon, CeliensIcon, CelientIcon, CeliepIcon, CelierIcon, CeliesIcon, CeliesaIcon, CeliesusIcon, CeliesyIcon, CelieusaIcon, CelieuxIcon, CelievaIcon, CeliexIcon, CelifexIcon, CeligeIcon, CeligesIcon, CeligexIcon, CeligheIcon, CeligherIcon, CeligherainIcon, CeligherainusIcon, CeligheriaIcon, CeligherusIcon, CeligheruxIcon, CelighesIcon, CelighexIcon, CelighextIcon, CeligheyIcon, CeligheysIcon, CelighezIcon, CelighezaIcon, CelighezusIcon, CelighezusaIcon, CeligheusIcon, CeligheusaIcon, CeligheusexIcon, CeligheuxIcon, CelighevaIcon, CelighevaexIcon, CelighexaIcon, CelighexasIcon, CelighexIcon, CeligheysIcon, CeligheyIcon, CelighezIcon, CelighezsIcon, CelighesIcon, CeligheaIcon, CeligheaexIcon, CeligheadIcon, CeligheadsIcon, ClientAccountHistoryTransactionDetails, ClientAccountHistoryTransactionDetailsExtended, ClientAccountListItem, ClientAccountListItemFocus, ClientAccountListItemHover, ClientAccountListItemPressed, ClientAccountProfileFullProfile, ClientAccountProfileFullProfileExtended, ClientAccountProfileMinimal, ClientAccountProfileMinimalExtended, ClientAccountProfileTransactionHistory, ClientAccountProfileTransactionHistoryExtended, ClientAccountSelectionListItem, ClientAccountSelectionListItemExtended, ClientAccountSelectionListItemFocus, ClientAccountSelectionListItemHover, ClientAccountSelectionListItemPressed, ClientAccountTransactionHistoryScreen, ClientAccountTransactionHistoryScreenExtended, ClientCreditProductsListItem, ClientCreditProductsListItemExtended, ClientCreditProductsScreen, ClientCreditProductsScreenExtended, ClientDashboard, ClientDashboardExtended, ClientDebitsAndCreditsScreen, ClientDebitsAndCreditsScreenExtended, ClientPaymentScreenSelectionScreen, ClientPaymentScreenSelectionScreenExtended, ClientProfileAccountHistoryScreen, ClientProfileAccountHistoryScreenExtended, ClientProfileFullProfile, ClientProfileFullProfileExtended, ClientProfileMinimal, ClientProfileMinimalExtended, ClientProfileTransactionDetailsScreen, ClientProfileTransactionDetailsScreenExtended, ClientProfileTransactionHistoryScreen, ClientProfileTransactionHistoryScreenExtended, ClientProfileTransactionScreen, ClientProfileTransactionScreenExtended, ClientTransactionDetailsScreen, ClientTransactionDetailsScreenExtended, CommonTransactionHistory, CommonTransactionHistoryExtended, CommonTransactionHistoryItem, CommonTransactionHistoryItemExtended, CommonTransactionHistoryItemFocus, CommonTransactionHistoryItemHover, CommonTransactionHistoryItemPressed, CommonTransactionItem, CommonTransactionItemExtended, CopyToClipboard, CountryFlag, CountrySelect, CreditCardEntryFieldIcon, CreditProductBadge, CreditProductIcon, CreditProductItem, CreditProductItemExtended, CreditProductItemFocus, CreditProductItemHover, CreditProductItemPressed, CurrencySelect, CustomCardNumberInput, CustomDateInput, CustomEmailInput, CustomNumberInput, CustomPasswordInput, CustomPhoneNumberInput, CustomSearchInput, CustomTextInput, CustomTimeInput, DashboardCardBalance, DashboardCardBalanceExtended, DashboardCardCreditLine, DashboardCardCreditLineExtended, DashboardCardPersonal, DashboardCardPersonalExtended, DashboardCardTransactionRecent, DashboardCardTransactionRecentExtended, DashboardScreen, DashboardScreenExtended, DatePicker, Delete, DeleteForever, DeleteIcon, DeleteOutline, DeleteSweep, DescriptionIcon, DesignServices, Desktop, DesktopMacIcon, DesktopWindowsIcon, DetailsIcon, DetectorIcon, DetailedDocumentIcon, DetailedDocumentIconExtended, DeveloperBoard, DeveloperBoardOff, DeveloperModeIcon, DevicesIcon, DewIcon, DiagnosisIcon, DialerSip, DialIcon, DialogIcon, DiamondsIcon, DiamondIcon, DiaphanousIcon, DiaphragmIcon, DiarrhoeaIcon, DiaryIcon, DibbleIcon, DibsIcon, DiceIcon, Dictaphone, DictateIcon, Dictionary, DictionaryEntryIcon, DidgeridooIcon, DiedIcon, DieIcon, DieselEngineIcon, DietIcon, DifferenceBetweenIcon, DifferenceIcon, DifferentIcon, DifferentiateIcon, Difficult, DifficultIcon, DiffuseIcon, DiffusionIcon, DiggingIcon, DigitalDisplayIcon, DigitalTwinIcon, DigitizeIcon, DiglotIcon, DignicIcon, DigressingIcon, DigressionIcon, DikeIcon, DilapidatedIcon, DilatableIcon, DilationIcon, DilateIcon, DilatorIcon, DilatorMuscleLabelIcon, DilatoryIcon, DilaterIcon, DilatorIcon, DilateIcon, DilationIcon, DilatorIcon, DilatorMuscleLabelIcon, DilatoryIcon, DiluteIcon, DilutePreviewIcon, DilutedIcon, DilutionIcon, DilutiveIcon, DimensionIcon, DimensionalIcon, DimensionalityIcon, DimensionsIcon, DimeIcon, DimeThickIcon, DimIcon, DimOutIcon, DimmedIcon, DimmediaIcon, DimmerIcon, DimmerSwitchIcon, DimmingIcon, DimlyIcon, DimmableIcon, DimmedIcon, DimmediaIcon, DimmerIcon, DimmerSwitchIcon, DimmingIcon, DimlyIcon

**Intentional Additions (3 custom components not in Figma kit):**
- AgentSignInFingerPrint — Biometric authentication for agent onboarding
- AgentSignInLoading — Loading state for agent sign-in flow
- SAMinorSAGuardian — South Africa–specific minor/guardian selection

All components ship as JSX + TypeScript interfaces and respond to Figma design tokens (color, spacing, typography, shadow). Single-color icons use `currentColor` for recoloring.

## Tokens

Figma Variables are exported as CSS custom properties organized by concern:

- **Color tokens**: `--color-primary`, `--color-error`, `--surface-container-low`, etc. (537 color variables)
- **Spacing tokens**: `--spacing-xs` through `--spacing-3xl` (224 float/dimension variables)
- **Typography tokens**: `--font-family-primary`, `--font-size-body-medium`, etc. (applied via fig-typography.css class rules)
- **Shadow tokens**: `--shadow-1`, `--shadow-2` (depth system)

Tokens support multiple theme modes (Light, Dark, Kendo breakpoints, IC variants, test modes). Components reference these tokens, so theme switching is as simple as toggling a mode class on the root element.

## Typography Styles

Generated CSS classes for every text style defined in the Figma file. Common styles include:

- `._display-large`, `._display-medium`, `._display-small` (hero/page-level headings)
- `._headline-large`, `._headline-medium`, `._headline-small` (section headings)
- `._title-large`, `._title-medium`, `._title-small` (subsection headings)
- `._body-large`, `._body-medium`, `._body-small` (reading copy)
- `._label-large`, `._label-medium`, `._label-small` (labels & captions)

## Starting Points

Consuming projects can use these components and screens as starting templates:

- **Component templates**: Button, Input, Avatar, Badge (in `components/` card pages)
- **Screen templates**: SA-Minor-SA-Guardian onboarding flow, Agent sign-in with biometric, Client profile view

## Project Structure

```
├── styles.css                        # Global entry point (imports tokens, typography, assets)
├── components/
│   ├── fig-tokens.css               # Figma variables as CSS custom properties
│   ├── fig-typography.css           # Text style classes
│   ├── fig-assets.css               # Image asset CSS classes
│   ├── assets/                      # Extracted image fills (67 bitmaps)
│   ├── *.jsx + *.d.ts               # Component source (175+ families)
│   └── screens/                     # Screen/template components
│       ├── SAMinorSAGuardian.jsx    # Onboarding flow
│       ├── AgentSignInFingerPrint.jsx
│       ├── ClientProfileFullProfile.jsx
│       └── ...
├── readme.md                        # This file
└── SKILL.md                         # CLI/code generation skill
```

## Known Limitations

- **Large images**: Two images (4.2 MiB each) exceeded per-asset budget and were skipped; their fills drop silently
- **Nested instance overrides**: Deep component nesting with per-character text overrides or complex instance swaps may not render exactly
- **Vector geometry**: A few boolean-operation shapes (CountryFlag component SVGs) could not be decoded and render as placeholder boxes; these are low-priority UI elements
- **Unresolved library aliases**: 2 token variables are library-aliased and not materialized; they are rarely used (check token usage before assuming a gap)
- **Text styles**: Figma's per-character style overrides and list markers are approximated; use the rendered components as reference

## Intentional Gaps

- **No logo**: The Figma file does not contain a brand mark. Render the name "Capitec" in Nunito Sans ExtraBold when a logo is needed; request a separate brand asset if a mark should be added
- **No icon pack CDN**: 143 icons are embedded in component fills. Use them via single-color recoloring in consuming projects

## Next Steps

1. **Verify in a consuming project**: Import this system into a new project and build a simple screen to confirm components render and tokens apply correctly
2. **Refine approximations**: Review extracted components (especially DocumentViewer, Complex forms) against the original Figma to fine-tune any baked-inline overrides
3. **Add @font-face rules**: Obtain Nunito Sans, Hind Vadodara, and Inter font files and add @font-face rules to `styles.css` so text renders in the correct typeface (fallback stacks are in place for now)
4. **Extend templates**: Add more screen templates (e.g., transaction history, settings, payment flows) as needs arise
5. **Test theme modes**: Verify Light/Dark/Kendo breakpoint mode switching works end-to-end in consuming projects

## Version & Attribution

- **Design system version**: 1.0 (extracted July 2026 from Figma)
- **Source**: Minor Onboarding.fig
- **Extracted by**: Capitec Design System compiler
- **Contact**: [Add team/org contact info]
