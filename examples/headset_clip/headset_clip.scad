include <BOSL2/std.scad>;

$fn = 100;                   // Default facet count for smooth circular geometry

// Render Quality Controls
// Fast low-res preview (F5) vs. smooth high-res final render (F6)
$fa = $preview ? 12 : 2;     // Minimum angle (degrees)
$fs = $preview ? 2 : 0.3;    // Minimum fragment size (mm)

// 3. BOSL2 Printer Tolerances
// Added clearance for 3D-printed joints, holes, and press-fits (in mm)
$slop = 0.2; 

// =================================================================
// CUSTOMIZER PARAMETERS
// =================================================================
/* [General Dimensions] */
length = 50;
width  = 30;
height = 20;

/* [Advanced Options] */
wall_thickness = 2.0;

// =================================================================
// DERIVED CONSTANTS & CALCULATIONS
// =================================================================
inner_length = length - (2 * wall_thickness);
inner_width  = width  - (2 * wall_thickness);

// =================================================================
// MODEL
// =================================================================
include <../openscad-lib/openscadgen-core.scad>;

// Use "horzSlice" or "vertSlice" to inspect a printable cross-section,
// "all" to show the object and both slice planes, or "obj" for the model.
renderType = "obj";

clipWidth = 25;
clipLength = 46;
clipThickness = 2;

barExtend = 30;

clipEndHeight = 10;
clipEndThickness = 2;

holeRadius = 2.5;
holeSeperate = 14.5;
holeOffset = 5;
holeFwd = -clipLength/2+holeOffset;

boltOffset = 3.5;



module headset_clip(){
difference(){
    cuboid([clipLength, clipWidth, clipThickness+clipEndHeight], rounding=3, anchor=BOT);
        move([clipEndThickness,1,clipThickness+0.001])
        cuboid([clipLength,clipWidth+6,barExtend+1], rounding=3, anchor=BOT);
        
        left(holeFwd)
        ycopies(holeSeperate, 2){
        zcyl(r=holeRadius, h=100);
           
//            
            }
        }
        up(clipThickness)
        left(holeFwd)
        ycopies(holeSeperate, 2){
         tube(ir=holeRadius, or=holeRadius+1, h=boltOffset, anchor=BOT);
         }
}

sliced(renderType=renderType) {
    headset_clip();
}

