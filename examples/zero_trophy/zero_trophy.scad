include <BOSL2/std.scad>;
include <../../openscad-lib/openscadgen-core.scad>;

$fn = 100;

// Use "obj" for the finished trophy, or either slice mode while sizing it.
renderType = "obj";

// "all" shows the complete trophy.  "base" is useful when checking the
// wording and engraving before committing to the top detail.
mode = "all";

/* [Print Output] */
printMode = "chairStandalone"; // [assembled,baseWithChairCutout,chairStandalone]

/* [Trophy] */
trophyMode = "campChair"; // [number,campChair]

baseWidth = 120;
baseDepth = 70;
baseHeight = 15;
wedgeHeight = 18;
flatTopDepth = 42;

/* [Engraving] */
baseText = "Bailey boys 2026";
baseText2 = "0 Goals, 0 Assists";
engravingDepth = 1.2;

/* [Zero] */
zeroHeight = 82;
zeroThickness = 10;
zeroFont = "Liberation Sans:style=Bold";

/* [Camp Chair] */
chairFrameWidth = 30;
chairFrameHeight = 90;
chairTubeWall = 3;
chairFrameThickness = 4;
chairTubeRounding = 1;
chairFrameTilt = 0;
clothThickness = 1.8;
clothSag = 1;
clothConnectorLift = 1.5;
clothCurveSegments = 20;
chairRotation = 120;
globalChairRotation =[-44,0,220];


globalChairMove =[-30,chairFrameHeight-10,baseHeight];

globalChairXRotation =-60;
rectTubeOneMove = [0, 0, 0];

module trophy_wedge() {
    // The wedge rises from the front of the base to the front side of the
    // raised cuboid. This leaves the cuboid's top as a substantial flat area.
    hull() {
        translate([0, -baseDepth/2, baseHeight])
            cuboid([baseWidth, 0.01, 0.01]);
        translate([0, -baseDepth/2 + baseDepth - flatTopDepth,
                   baseHeight + wedgeHeight/2])
            cuboid([baseWidth, 0.01, wedgeHeight]);
    }
}

module sloped_wedge_engraving(label, textSize, position) {
    wedgeRun = baseDepth - flatTopDepth;
    slopeAngle = atan(wedgeHeight / wedgeRun);

    // Rotate the text's plane onto the wedge's upward-facing slope. Starting
    // its extrusion just inside the surface creates a recessed, readable mark.
    translate([0, -baseDepth/2, baseHeight])
        rotate([slopeAngle, 0, 0])
            translate([0, wedgeRun * position, -engravingDepth])
                linear_extrude(height=engravingDepth + 0.02)
                    text(label, size=textSize, font="Liberation Sans:style=Bold",
                         halign="center", valign="center");
}

module zero_tube() {
    // The counter in the glyph makes this an actual 0-shaped tube rather than
    // a solid disc. It is attached to the rear of the flat cuboid top.
    translate([0, baseDepth/2 - zeroThickness, baseHeight + wedgeHeight])
        rotate([-90, 0, 0])
            mirror([0, 1, 0])
                linear_extrude(height=zeroThickness)
                    text("0", size=zeroHeight, font=zeroFont,
                         halign="center", valign="bottom");
}

module sqrTube(size, wall, rounding) {
    // A rounded outer cuboid minus an inner cuboid makes a locally owned,
    // square/rectangular tube. Negative rounding on the cutout keeps the
    // inner corners complementary to the outer edge treatment.
    assert(size[0] > 2 * wall && size[1] > 2 * wall);
    difference() {
        cuboid(size, rounding=rounding);
        cuboid([size[0] - 2 * wall, size[1] - 2 * wall, size[2] + 0.02],
               rounding=-rounding);
    }
}

module camp_chair_frame(tilt) {
    // Both tubes share the same height. Different opposite rotations make
    // them overlap as a folding chair's X-shaped frame.
    translate([0, 0, baseHeight + wedgeHeight + chairFrameHeight/2])
        rotate([tilt, 0, 0])
            rotate([90, 0, 0])
            sqrTube([chairFrameWidth, chairFrameHeight, chairFrameThickness],
                    chairTubeWall, chairTubeRounding);
}

function cloth_curve_z(t, startZ, endZ, sag) =
    let(middleZ = min(startZ, endZ) - sag)
    t <= 0.5
        ? startZ + (middleZ - startZ) * (4 * t - 4 * t * t)
        : middleZ + (endZ - middleZ) * (2 * t - 1) * (2 * t - 1);

function frame_top_rail_y(tilt) =
    -sin(tilt) * (chairFrameHeight/2 - chairTubeWall/2);

function frame_top_rail_z(tilt) =
    baseHeight + wedgeHeight + chairFrameHeight/2
    + cos(tilt) * (chairFrameHeight/2 - chairTubeWall/2);

function cloth_profile_points(startY, endY, upperZ, lowerZ, sag, thickness, segments) =
    concat(
        [for (index = [0 : segments])
            let(t = index / segments,
                y = startY + (endY - startY) * t,
                z = cloth_curve_z(t, upperZ, lowerZ, sag)) [-z, y]],
        [for (index = [segments : -1 : 0])
            let(t = index / segments,
                y = startY + (endY - startY) * t,
                z = cloth_curve_z(t, upperZ, lowerZ, sag) - thickness) [-z, y]]
    );

module camp_chair_cloth() {
    // Place each end at the centre of its actual rotated top rail, giving the
    // extruded cloth enough overlap to fuse with both rectangular tubes.
    lessTiltTopY = frame_top_rail_y(-chairFrameTilt) + rectTubeOneMove[1];
    moreTiltTopY = frame_top_rail_y(chairRotation);
    lessTiltTopZ = frame_top_rail_z(-chairFrameTilt)
                  + rectTubeOneMove[2] + clothConnectorLift;
    moreTiltTopZ = frame_top_rail_z(chairRotation) + clothConnectorLift;
    clothRailHalfWidth = chairFrameWidth/2;// - chairTubeWall;
    clothMinX = min(-clothRailHalfWidth,
                    rectTubeOneMove[0] - clothRailHalfWidth);
    clothMaxX = max(clothRailHalfWidth,
                    rectTubeOneMove[0] + clothRailHalfWidth);

    // Two smooth quadratic halves meet at a deliberate centre trough, so
    // unequal attachment heights cannot pull the cloth's low point sideways.
    translate([(clothMinX + clothMaxX)/2, 0, 0])
        rotate([0, 90, 0])
            linear_extrude(height=clothMaxX - clothMinX, center=true)
                polygon(points=cloth_profile_points(
                    moreTiltTopY, lessTiltTopY, moreTiltTopZ, lessTiltTopZ, clothSag,
                    clothThickness, clothCurveSegments
                ));
}

module camp_chair() {
    translate(rectTubeOneMove)
        camp_chair_frame(-chairFrameTilt);
    camp_chair_frame(chairRotation);
    camp_chair_cloth();
}

module placed_camp_chair() {
    // Keep assembled, cutout, and standalone chair geometry in exactly the
    // same coordinate system so the two separately printed parts match.
    move(globalChairMove)
        rotate(globalChairRotation)
                camp_chair();
}

module standalone_camp_chair() {
    // A standalone print does not need the trophy's placement transform. Keep
    // its geometry identical while centring it for predictable slicer/camera use.
    move([0, 0, -(baseHeight + wedgeHeight + chairFrameHeight/2)])
        camp_chair();
}

module trophy_top() {
    if (trophyMode == "number") {
        zero_tube();
    } else if (trophyMode == "campChair") {
        placed_camp_chair();
    } else {
        assert(false, str("Unknown trophyMode: ", trophyMode));
    }
}

module trophy_base(withChairCutout=false) {
    difference() {
        union() {
            cuboid([baseWidth, baseDepth, baseHeight], anchor=BOTTOM);
            translate([0, baseDepth/2 - flatTopDepth/2, baseHeight + wedgeHeight/2])
                cuboid([baseWidth, flatTopDepth, wedgeHeight]);
            trophy_wedge();
        }

        sloped_wedge_engraving(baseText, 10, 0.67);
        sloped_wedge_engraving(baseText2, 9, 0.34);
        if (withChairCutout) placed_camp_chair();
    }
}

module zero_trophy() {
    if (trophyMode == "campChair" && printMode == "chairStandalone") {
        standalone_camp_chair();
    } else {
        union() {
            trophy_base(trophyMode == "campChair"
                        && printMode == "baseWithChairCutout");
            if (mode == "all" && printMode == "assembled") trophy_top();
        }
    }
}

sliced(renderType=renderType) zero_trophy();
