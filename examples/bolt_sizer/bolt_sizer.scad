
// 1. Library Includes
include <BOSL2/std.scad>

$preview = true;
// 2. Render Quality Controls
// Fast low-res preview (F5) vs. smooth high-res final render (F6)
$fa = $preview ? 12 : 2;     // Minimum angle (degrees)
$fs = $preview ? 2 : 0.3;    // Minimum fragment size (mm)
$fn = 20;                     // Disabled so $fa and $fs take effect

// 3. BOSL2 Printer Tolerances
// Added clearance for 3D-printed joints, holes, and press-fits (in mm)
$slop = 0.2;

// =================================================================
// CUSTOMIZER PARAMETERS
// =================================================================

/* [Render Controls] */
// Select the print-oriented export: box, bottomColorModifier,
// topColorModifier, modifierV2, preview, or test.
mode = "box";

// The export bed plane.  "right-side-up" keeps the sizer's Z direction; a
// modifier then includes a thin spacer down to Z=0 so Bambu Studio preserves
// its intended height when it drops each imported STL to the floor.
exportOrientation = "right-side-up"; // [on-side,right-side-up]

/* [Sizer Body] */
// Overall body dimensions and the pointed end attached to it.
globalSizerHeight = 10;
sizerRounding = 1;
bottomGapSize = 16;

pointHoleDown = 10;
pointBottomHeight = 12;
pointFwdOffset = 0;

/* [Bolt Holes] */
// Repeated bolt-hole cut-outs along the length of the sizer.
boltHoleWidth = 15;
boltHoleGap = 2;
boltHoleCount = 6;
// The solid base beneath the bolt holes and top labels.
floorHeight = 2;
boltHoleHeight = 8;
boltHoleSideWallSize = 3;
holeDepth = 12;

/* [Golf Pick Hole] */
// Long relief cut and its larger end clearance.
pickHoleDown = 3;
pickRadius = 5.2;
pickHoleIn = -floorHeight + 2;
pickChamfer = 5;
pickBottomRadius = 3.5;
pickBottomHeight = 18;

/* [Top Labels] */
// Depth of the recessed labels below the floor surface.
textIndentHeight = 0.3;
// One thin, full-footprint colour cuboid overlaps the floor of the numbered
// recesses by this amount. It must remain below the top section, leaving only
// the layer's upper face and perimeter sides visible inside each cutout.
topColorModifierCutoutOverlap = 0.01;
topColorModifierHeight = 0.1;
topColorModifierStripSize = 0.1;
textSize = 5.5;
textShift = 2.5;

/* [Bottom Image Relief] */
// Height-map relief generated from image.png on the bottom of the sizer.
bottomRenderTextHeight = 1;
bottomRenderModifierHeight = -2.15;
bottomRenderFloorOffset = 0.4;
bottomReliefLayerSpacing = 0.05;
bottomReliefLayerCount = 10;
// Keep the inverted cavity aligned to the separately exported modifier.
bottomColorCutoutInset = 0;
// How far the logo cavity enters the bottom face.  The separate modifier
// continues farther into the floor, so the imported parts still overlap.
bottomColorCutoutDepth = 0.2;
// Extend each separate colour modifier this far into the remaining box solid.
// That overlap makes the parts register as one object in a multi-part slicer.
modifierModelOverlap = 0.15;

// Simple top/bottom modifier placement. The small overhang clears the model
// sides. It enters each text recess enough to give the printed colours a
// common layer, while leaving an opaque outer skin over the rest of each face.
modifierV2SideOverhang = -0.5;
modifierV2BottomTextOverlap = 0.05;
modifierV2TopTextSurfaceClearance = 0.05;
modifierV2SpacerSize = 0.2;

logoScaleAdjust = 2.9;
logoShiftRight = 0;
// Rotate the landscape source image into its readable bottom-face orientation.
logoImportRotation = 180;
logoImportYRotation = 180;


// =================================================================
// DERIVED CONSTANTS & CALCULATIONS
// =================================================================
globalSizerLenght = (boltHoleHeight + boltHoleGap) * boltHoleCount + boltHoleGap;
globalSizerSize = [boltHoleWidth + boltHoleSideWallSize, globalSizerLenght, globalSizerHeight];

// Pointed end geometry, derived from the body and bolt-hole dimensions.
bottomPointRadius = 21;
bottomPointBaseAngleSpread = boltHoleWidth + 11;
bottomPointUpShift = 11.4;
bottomPointMove = [0, bottomPointUpShift - bottomPointRadius, 0];




// =================================================================
// MODEL
//


	/*
	renderType:
	use to print a test slice and confirm sizing before printing:
	 - "horzSlice" - horizontal slices (default)
	 - "vertSlice" - vertical slices
	 - "all" - the whole object
	*/







    // Label recesses are defined directly from the top of the solid floor.
    module labelText(height = textIndentHeight, indent = textIndentHeight) {
        ycopies(boltHoleHeight + boltHoleGap, n = boltHoleCount)
            fwd(textShift)
                up(floorHeight - indent)
                    text3d(str(($idx * 10) + 20), h = height, anchor = BOT,
                        size = textSize, font = ":style=bold");
    }

    // BETA: one continuous, thin colour sheet behind all top number cutouts.
    // Its small grounding strip reaches the common import plane without
    // changing the sheet's position inside the recesses.
module topColorModifier() {
        modifierBottom = floorHeight - textIndentHeight - topColorModifierCutoutOverlap;
        modifierTop = modifierBottom + topColorModifierHeight;
        assert(modifierTop <= floorHeight,
            "topColorModifier must stay inside the top label recess.");
        union() {
            up(modifierBottom)
                cuboid([globalSizerSize[0], globalSizerLenght, topColorModifierHeight], anchor = BOT);
            right(globalSizerSize[0] / 2 - topColorModifierStripSize / 2)
                cuboid([topColorModifierStripSize, topColorModifierStripSize, modifierBottom], anchor = BOT);
        }
    }


    module roundedHole(holeHeight = boltHoleHeight, holeWidth = boltHoleWidth){
    cuboid([holeWidth,holeHeight,holeDepth/2], anchor=BOT);
    //, rounding=3, edges=BOTTOM, except=BACK
            up(holeDepth/2)
                       cuboid([holeWidth,holeHeight,holeDepth/2], anchor=BOT); //, rounding=-2, edges=TOP, except=BACK
    }

    module bolt_holes(boltHoleWidth=boltHoleWidth){
        ycopies(boltHoleHeight + boltHoleGap, n = boltHoleCount){
        if($idx==0){
        down(0)
        fwd(pointHoleDown)
        intersection(){
       // scale([1,1,2])
            // half_of([0, -6, 0], cp = -6){

           bottom_point(pointBottomHeight = pointBottomHeight * 1.7);
        //    }
                roundedHole(holeHeight = boltHoleHeight + 2, holeWidth = boltHoleWidth);


        }
        }

       roundedHole(holeHeight = boltHoleHeight, holeWidth = boltHoleWidth);


            }
    }

    module golf_pick_hole(){
        up(globalSizerHeight-pickHoleIn)
        fwd(pickHoleDown)
        ycyl(r=pickRadius, h=globalSizerLenght-pickBottomHeight/2+bottomGapSize/2, chamfer1=pickChamfer);
        //down(globalSizerLenght+bottomGapSize/2+5)
        if(pickBottomHeight > 0){
//            up(globalSizerHeight-pickHoleIn)
            up(globalSizerHeight)
            fwd((globalSizerLenght+bottomGapSize/2)/2)
             ycyl(r=pickBottomRadius, h=pickBottomHeight);
         }
    }

    module bottom_point(pointBottomHeight=pointBottomHeight){
    difference(){
        move(bottomPointMove)
        rotate([0,0,90])
         linear_extrude(h=globalSizerHeight-0.5)
        arc(r=bottomPointRadius, angle=[-bottomPointBaseAngleSpread, bottomPointBaseAngleSpread], rounding=[1, 1, 1], wedge=true);

        fwd(pointBottomHeight)
        down(0.001)
           cuboid([50,20,20], anchor=BOT+CENTER);
        }
    }

module logoProjection(height = 10, scaleZ = logoScaleAdjust) {
    rotate([0, 180, 90])
        resize([0, 0, scaleZ], auto = [true, true, false]) {
            // Rotate in the logo's 2D plane; doing this before the outer
            // orientation preserves the relief's Z placement and clipping.
            rotate([0, 0, logoImportRotation])
                // A Y flip reverses extrusion direction, so lift it by its
                // height to keep the relief in the same clipped Z range.
                up(height)
                    rotate([0, logoImportYRotation, 0])
                        linear_extrude(height = height, convexity = 10) {
                            projection(cut = true) {
                                translate([0, 0, 80]) {
                                    surface(file = "image.png", center = true, invert = true);
                                }
                            }
                        }
        }
}

// Image-relief modifier, clipped to the separate lower section. The source
// surface represents the background, so subtract it from the clip to retain
// the logo itself as the cavity and colour-modifier geometry.
module bottomRender(textHeight = 10, reliefDepth = 5, logoScaleAdjust = logoScaleAdjust) {
    up(bottomRenderModifierHeight)
        zcopies(bottomReliefLayerSpacing, n = bottomReliefLayerCount) {
            fwd(logoShiftRight)
                difference() {
                    // The image surface is projected below Z=0. A TOP-anchored
                    // clip volume expresses that directly, without offsets.
                    down(-0.001)
                    cuboid([globalSizerSize[0], globalSizerSize[1], floorHeight],
                        anchor = TOP);
                    logoProjection(height = textHeight, scaleZ = logoScaleAdjust);
                }
        }
}

// Position the relief to overlap the bottom slice before it is flipped.
module bottomRenderAtFloor() {
    up(bottomRenderFloorOffset + 0.4)
        bottomRender(textHeight = bottomRenderTextHeight + 10);
}

// Flip the relief for the bottom face.  Both the standalone colour modifier
// and the cavity in the box use this same orientation.
module invertedBottomLogoRelief() {
    rotate([180, 0, 0])
        bottomRenderAtFloor();
}


// The three sections together make one complete modifier height.  The logo
// itself is the bottom section; the cuboids begin at its upper face.
// This applies only to the standalone modifier.  The box continues to use
// invertedBottomLogoRelief() at its existing cutout depth.
limitedLogoForModifierHeight = 0.3;
midModiferSectionHeight = 0.4;
topModiferSectionHeight = 0.3;

// Reach into the top frame on both axes so the middle and top sections
// overlap instead of merely meeting at an edge.
midModiferSectionScaleX = 0.7;
midModiferSectionScaleY = 0.9;

topModiferSectionScaleX = 0.5;
topModiferSectionScaleY = 1.4;
// Use nearly all of the recess available below the unchanged modifier top,
// making the printed text taller without entering the box's outer surface.
topModifierTextShare = 0.33;
topModifierFrameOutsideOverhang = 1;
topModifierFrameWall = 2;
// Adjacent modifier pieces overlap slightly, producing one watertight part
// rather than separate shells that only share a face.
modifierConnectionOverlap = 0.01;

bottomModiferReduceX = 3.4;
bottomModiferReduceY = 3;

// Extend the modifier 0.4 mm farther below the floor while retaining its
// existing upper Z boundary.
floorModifierStartHeight = -0.2;
floorModifierTotalHeight = 2.1;

// The logo is clipped to the height allocated to the modifier's bottom
// section, giving it a flat top face for the first cuboid to attach to.
module limitedLogoForModifier() {
    intersection() {
        cuboid([
            globalSizerSize[0] - bottomModiferReduceX,
            globalSizerSize[1] - bottomModiferReduceY,
            floorModifierTotalHeight * limitedLogoForModifierHeight
        ], anchor = BOT);
        down(3.2)
            invertedBottomLogoRelief();
    }
}

// Reuse the box's number layout through the full height of its recess.
// Its upper face is exactly the modifier's existing upper Z boundary.
module topModifierNumberStack() {
    modifierTop = floorModifierStartHeight + floorModifierTotalHeight;
    numberHeight = textIndentHeight * topModifierTextShare;
    numberBottom = modifierTop - numberHeight;
    assert(modifierTop > floorHeight - textIndentHeight && modifierTop <= floorHeight,
        "The modifier top must lie within the top number recesses.");
    assert(numberBottom >= floorHeight - textIndentHeight,
        "The modifier number layer must remain within the top number recesses.");
    ycopies(boltHoleHeight + boltHoleGap, n = boltHoleCount)
        fwd(textShift)
            up(floorModifierTotalHeight - numberHeight)
                text3d(str(($idx * 10) + 20), h = numberHeight, anchor = BOT,
                    size = textSize, font = ":style=bold");
}

// A frame, rather than a solid plate, preserves the visible outer modifier
// impression without covering the number stack in the middle.
module topModifierFrame() {
fwd(5)
scale([1,0.89,1])
    rect_tube(
        size = [
            globalSizerSize[0] + 2 * topModifierFrameOutsideOverhang,
            globalSizerSize[1] + 2 * topModifierFrameOutsideOverhang + 20
        ],
        wall = topModifierFrameWall,
        h = floorModifierTotalHeight * topModiferSectionHeight + modifierConnectionOverlap,
        anchor = BOT
    )
    attach(FWD){
//fwd(4)
        cuboid([globalSizerSize[0],floorModifierTotalHeight * topModiferSectionHeight + modifierConnectionOverlap, 13], anchor=TOP);
        }
    
}

// Each rail overlaps the perimeter frame and its corresponding text. Together
// with the enlarged middle section this makes the whole modifier continuous.
module topModifierTextSupports() {
    frameInnerSize = [
        globalSizerSize[0] + 2 * topModifierFrameOutsideOverhang - 2 * topModifierFrameWall,
        globalSizerSize[1] + 2 * topModifierFrameOutsideOverhang - 2 * topModifierFrameWall
    ];
    ycopies(boltHoleHeight + boltHoleGap, n = boltHoleCount)
        fwd(textShift)
            cuboid([
                frameInnerSize[0] + 2 * modifierConnectionOverlap,
                textSize + 2 * modifierConnectionOverlap,
                floorModifierTotalHeight * topModiferSectionHeight + modifierConnectionOverlap
            ], anchor = BOT);
}

module bottomLogoModifier() {
    sectionHeightTotal = limitedLogoForModifierHeight +
        midModiferSectionHeight + topModiferSectionHeight;
    assert(abs(sectionHeightTotal - 1) < 0.000001,
        "limitedLogoForModifierHeight + midModiferSectionHeight + topModiferSectionHeight must equal 1");

    up(floorModifierStartHeight)
        union() {
            limitedLogoForModifier();

            // limitedLogoForModifier() is an intersection, which has no
            // BOSL attachment metadata.  Use its declared height instead, so
            // each cuboid is retained and begins exactly where its section
            // below ends.
            up(floorModifierTotalHeight * limitedLogoForModifierHeight -
                modifierConnectionOverlap)
                cuboid([
                    globalSizerSize[0] * midModiferSectionScaleX,
                    globalSizerSize[1] * midModiferSectionScaleY,
                    floorModifierTotalHeight * midModiferSectionHeight +
                        modifierConnectionOverlap
                ], anchor = BOT);

            up(floorModifierTotalHeight *
                (limitedLogoForModifierHeight + midModiferSectionHeight) -
                modifierConnectionOverlap)
                union() {
                    topModifierFrame();
                    //topModifierTextSupports();
                }
               zcopies(0.1,5)
            topModifierNumberStack();
        }
}

// The shared middle layer reaches both the top label recess and bottom logo
// recess.  It deliberately overhangs the sides so it remains visible and easy
// to select in a multi-part slicer.
function modifierV2Bottom() = bottomColorCutoutDepth - modifierV2BottomTextOverlap;
function modifierV2Top() = floorHeight - modifierV2TopTextSurfaceClearance;

module modifierV2MiddleLayer() {
    modifierBottom = modifierV2Bottom();
    modifierTop = modifierV2Top();
    assert(modifierBottom > 0 && modifierBottom < bottomColorCutoutDepth,
        "Middle modifier must enter the bottom text recess without reaching the bottom surface.");
    assert(modifierTop > floorHeight - textIndentHeight && modifierTop < floorHeight,
        "Middle modifier must reach the top text recess without reaching the top surface.");

    up(modifierBottom)
        cuboid([
            globalSizerSize[0] + modifierV2SideOverhang * 2,
            globalSizerSize[1] + modifierV2SideOverhang * 2,
            modifierTop - modifierBottom
        ], anchor = BOT);
}

// The complete V2 colour modifier: one simple middle layer plus the printable
// portion of bottomRender.  Keeping the bottom relief here makes this export
// self-contained while still sharing its geometry with the box cutout.
module modifierV2() {
    union() {
     //   modifierV2MiddleLayer();
        bottomLogoModifier();
    }
}

// The normal box export is the complete assembled sizer.  Keep the individual
// modifier modes below for multi-colour slicer imports, but do not make the
// default box STL silently omit those model volumes.
module assembledBox() {
    union() {
        boxWithBottomColorCutout();
       // modifierV2()
        //#topColorModifier();
      //  bottomLogoModifier();
    }
}

// Image-only assembly view.  The printable box and modifier remain separate
// exports; this scene makes the V2 insert visible in the generated close-ups.
module modifierV2Preview() {
    color("grey")
        assembledBox();
    color("red")
        modifierV2();
}

// A deliberately small placement datum, outside the body footprint. It is
// only emitted upright, where Bambu Studio otherwise lowers a floating middle
// layer to the bed and loses its designed Z offset.
module modifierV2Spacer() {
    right(globalSizerSize[0] / 2 + modifierV2SideOverhang - modifierV2SpacerSize / 2)
        back(globalSizerSize[1] / 2 + modifierV2SideOverhang - modifierV2SpacerSize / 2)
            cuboid([
                modifierV2SpacerSize,
                modifierV2SpacerSize,
                modifierV2Bottom()
            ], anchor = BOT);
}

// The cavity is deliberately shallow, leaving the rest of the modifier
// embedded in the floor so the separately imported logo joins to the box.
module bottomLogoCutout() {
    bottomModifierTop = bottomColorCutoutInset -
        (bottomRenderFloorOffset + 0.4 + bottomRenderModifierHeight) +
        (bottomReliefLayerCount - 1) * bottomReliefLayerSpacing / 2;
    intersection() {
        bottomLogoModifier();
        cuboid([globalSizerSize[0], globalSizerSize[1],
            min(bottomColorCutoutDepth, bottomModifierTop - modifierModelOverlap)],
            anchor = BOT);
    }
}

// The bottom is an independent cuboid.  Its top edges stay square so it joins
// the upper box flush; the visible lower edges retain the body rounding.
module bottomSection() {
    cuboid([globalSizerSize[0], globalSizerSize[1], floorHeight], anchor = BOT,
        rounding = sizerRounding, except_edges = TOP);
}

module boxSection() {
    up(floorHeight)
        cuboid([globalSizerSize[0], globalSizerSize[1], globalSizerHeight - floorHeight],
            anchor = BOT, rounding = sizerRounding, except_edges = BOT);
}

// Main sizer body with the bottom-colour relief cut into its base.
module boxWithBottomColorCutout() {
    difference() {
        boltSizerBody();
        up(bottomColorCutoutInset)
        down(3.5)
        //    bottomLogoCutout();
            
            invertedBottomLogoRelief();
    }
}

// Put an export on its long side with the same bed plane.
module printBoxOnSide() {
    up(globalSizerSize[0] / 2)
        rotate([0, 90, 0])
            children();
}

module printExport(includePlacementSpacer = false) {
    assert(exportOrientation == "on-side" || exportOrientation == "right-side-up",
        str("Unknown exportOrientation: ", exportOrientation));
    if (exportOrientation == "on-side") {
        printBoxOnSide()
            children();
    } else {
        if (includePlacementSpacer) {
            union() {
                children();
                modifierV2Spacer();
            }
        } else {
            children();
        }
    }
}

module boltSizerBody() {
    difference() {
            union() {
                bottomSection();
                boxSection();
                fwd(globalSizerLenght / 2 + bottomGapSize / 2 - pointFwdOffset)
                    bottom_point();
            }
            up(floorHeight)
                bolt_holes();
            labelText();
            golf_pick_hole();
    }
}

function rangesOverlap(firstBottom, firstTop, secondBottom, secondTop) =
    firstTop > secondBottom && secondTop > firstBottom;

// Locks in the geometry that the directional test images exercise. The top
// sheet must overlap the bottom of the number recess without reaching the top
// section, and the bottom relief must enter (but not pass through) the solid
// floor after the cutout inset is applied.
module assertTestHighlighting() {
    topRecessBottom = floorHeight - textIndentHeight;
    topRecessTop = floorHeight;
    topModifierBottom = topRecessBottom - topColorModifierCutoutOverlap;
    topModifierTop = topModifierBottom + topColorModifierHeight;
    // After flipping and lowering into the bottom face, this is the upper
    // edge of the relief retained within the floor.
    bottomCutoutTop = bottomColorCutoutInset -
        (bottomRenderFloorOffset + 0.4 + bottomRenderModifierHeight) +
        (bottomReliefLayerCount - 1) * bottomReliefLayerSpacing / 2;
    bottomCavityTop = min(bottomColorCutoutDepth,
        bottomCutoutTop - modifierModelOverlap);

    assert(
        topModifierBottom < topRecessBottom && topModifierTop > topRecessBottom &&
            topModifierTop < topRecessTop,
        "Top colour sheet must overlap the recess floor without reaching the top section."
    );
    assert(
        bottomCavityTop > 0 && bottomCavityTop < bottomCutoutTop &&
            bottomCutoutTop <= floorHeight &&
            rangesOverlap(bottomCavityTop, floorHeight, 0, bottomCutoutTop),
        "Bottom colour cutout must be shallower than its modifier and stay within the floor."
    );
}

// Test-only scene: the printable exports remain separate, while both colour
// modifiers are red so the left/right image checks can assert they are visible.
module testExports() {
    assertTestHighlighting();
    printBoxOnSide()
        boxWithBottomColorCutout();
    left(globalSizerSize[0] + 8)
        printBoxOnSide()
            color("red")
                topColorModifier();
    right(globalSizerSize[0] + 8)
        printBoxOnSide()
            color("red")
                bottomLogoModifier();
}

// Unit-test scene: only the volumes common to each modifier and the completed
// box are emitted. If either overlap disappears, its coloured geometry does
// too; assertTestHighlighting() also fails the configured render immediately.
module modifierOverlapTest() {
    assertTestHighlighting();
    color("red")
        intersection() {
            boxWithBottomColorCutout();
            topColorModifier();
        }
    color("blue")
        intersection() {
            boxWithBottomColorCutout();
            bottomLogoModifier();
        }
}

if (mode == "box") {
//front_half()
    printExport()
        assembledBox();
//} else if (mode == "bottomColorModifier") {
 //   printExport()
  //      bottomLogoModifier();
} else if (mode == "topColorModifier") {
    printExport()
        topColorModifier();
} else if (mode == "modifierV2") {
    //printExport(includePlacementSpacer = true)
        bottomLogoModifier();
} else if (mode == "preview") {
    printExport()
        modifierV2Preview();
} else if (mode == "previewCut") {
front_half()
    printExport()
        modifierV2Preview();
} else if (mode == "test") {
    testExports();
} else if (mode == "overlapTest") {
    modifierOverlapTest();
} else {
    assert(false, str("Unknown mode: ", mode));
}
