include <BOSL2/std.scad>;
include <BOSL2/screws.scad>;

$fn = 200;                   // Default facet count for smooth circular geometry

// Render Quality Controls
// Fast low-res preview (F5) vs. smooth high-res final render (F6)
$fa = $preview ? 12 : 20;     // Minimum angle (degrees)
$fs = $preview ? 20 : 0.3;    // Minimum fragment size (mm)

// 3. BOSL2 Printer Tolerances
// Added clearance for 3D-printed joints, holes, and press-fits (in mm)
$slop = 0.2; 

// =================================================================
// CUSTOMIZER PARAMETERS
// =================================================================
/* [General Dimensions] */
type = "nut"; // nut | bolt | nut-or-bolt
spec = "15/64"; // [spec,length] M6,1

screwLength = 30;
/* [Advanced Options] */
// =================================================================
// MODEL
// =================================================================
include <../openscad-lib/openscadgen-core.scad>;

// Use "horzSlice" or "vertSlice" to inspect a printable cross-section,
// "all" to show the object and both slice planes, or "obj" for the model.
renderType = "obj";

module any_bolt(){
    if(type == "nut"){
      nut(spec,thickness=thickness);
    } else if(type == "bolt"){
      screw(str(spec,",",screwLength), head="hex", tolerance="6g", anchor=BOTTOM);
    } else if (type == "nut-and-bolt"){
        xdistribute(spacing=22){
          nut(spec);
            
          
               
      //     screw(str(spec,",",screwLength), head="hex", tolerance="6g", anchor=BOTTOM);
       }
    } else {
       assert(false, str("could not find type: ", type));
    }
    
}

sliced(renderType=renderType) {
    any_bolt();
}

