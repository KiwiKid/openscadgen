include <BOSL2/std.scad>;
include <BOSL2/screws.scad>;

$fn = 50;

partType = "holder"; // "holder" or "cordPlug"
mountMode = "wall"; // "wall" or "wedge"
device = "DefaultPhoneWithThickCase";
wallSize = 3;
bottomWallSize = 15;

yScaleFactor = 1.2;
xScaleFactor = 2.5;
defaultPhoneWithThickCaseSize = [158.2, 76.7*1.2, 8.2 * xScaleFactor];
holderRounding = 6;

cordRadius = 25;
cordDepth = 5;
cordHoleHeight = 60;
plugBodySize = [20, 35, 10];
cordHoleDown = 8;
cordHoleDepress = 15;

cordHoleMove = [cordHoleDown, 0, cordHoleDepress];

cordHoleOffsetMove = [0,0,defaultPhoneWithThickCaseSize[2]/2];

cordPlugThickness = 1.5;
cordPlugClearance = 0.15;

mountHoleDepth = 16;

screwSize = "M3";
screwHoleY = 25;
screwHoleLeftX = -defaultPhoneWithThickCaseSize[0] / 2 + 10;
screwHoleRightX = defaultPhoneWithThickCaseSize[0] / 2 - bottomWallSize - 30;
screwHolePositions = [
    [screwHoleLeftX, -screwHoleY],
    [screwHoleLeftX, screwHoleY],
    [screwHoleRightX, 0],
];

module wallMountScrewHoles() {
    for (position = screwHolePositions) {
        translate([position[0], position[1], wallSize])
        screw_hole(screwSize, head="flat", counterbore=0, length=wallSize + 1, anchor=TOP, $slop=0.1);
    }
}

module cordHole() {
     move(cordHoleMove)
    cuboid(plugBodySize, anchor=DOWN, rounding=1)
    attach(FWD)
    cuboid([cordRadius, cordDepth, cordHoleHeight], anchor=BOT+BACK, rounding=1);
}

// Thin cover that follows the cord hole opening and sits flush with the
// underside of the holder.
module cordPlug() {
    plugX = defaultPhoneWithThickCaseSize[0] / 2;
    holderBottomX = (defaultPhoneWithThickCaseSize[0] + bottomWallSize) / 2 - holderRounding;
    holderBottomFrontY = -(defaultPhoneWithThickCaseSize[1] + wallSize) / 2 + holderRounding;
    holeBackY = plugBodySize[1] / 2;
    plateLeftX = plugX - plugBodySize[0] / 2;
    plateRightX = min(plugX + plugBodySize[0] / 2, holderBottomX);
    plateCenterX = (plateLeftX + plateRightX) / 2;
    plateCenterY = (holderBottomFrontY + holeBackY) / 2;
    plateCenterZ = cordPlugThickness / 2;
    sliceThickness = 0.1;

    translate([-plateCenterX, -plateCenterY, -plateCenterZ])
    linear_extrude(height=cordPlugThickness)
    offset(delta=-cordPlugClearance)
    projection(cut=false)
    intersection() {
        right(defaultPhoneWithThickCaseSize[0] / 2 - cordHoleDown)
        down(cordHoleDepress)
        cordHole();

        cuboid(defaultPhoneWithThickCaseSize + [bottomWallSize, wallSize, 0], anchor=DOWN, rounding=holderRounding);

        translate([0, 0, sliceThickness / 2])
        cuboid([1000, 1000, sliceThickness], anchor=CENTER);
    }
}

module objectHolder(device = "DefaultPhoneWithThickCase") {
    if (device == "DefaultPhoneWithThickCase") {
        difference() {
            cuboid(defaultPhoneWithThickCaseSize + [bottomWallSize, wallSize, 5], anchor=DOWN, rounding=holderRounding);

            union() {
                up(wallSize)
                left(bottomWallSize / 2+1)
                cuboid(defaultPhoneWithThickCaseSize, anchor=DOWN, rounding=holderRounding);
                
                up(defaultPhoneWithThickCaseSize[2])
                left(mountHoleDepth+6)
                cuboid([defaultPhoneWithThickCaseSize[0]*0.9, defaultPhoneWithThickCaseSize[1]*1.1, defaultPhoneWithThickCaseSize[2]], anchor=DOWN, rounding=holderRounding);

                right(defaultPhoneWithThickCaseSize[0] / 2 - cordHoleDown)
                down(cordHoleDepress)
                move(cordHoleOffsetMove)
               
                cordHole();

                if (mountMode == "wall") {
                    wallMountScrewHoles();
                }
            }
        }
    }
}

module charging_cord_integrated_electronics_holder() {
    objectHolder(device=device);
}

if (partType == "cordPlug") {
    cordPlug();
} else {
    charging_cord_integrated_electronics_holder();
}
