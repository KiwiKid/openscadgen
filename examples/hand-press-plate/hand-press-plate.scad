
include <BOSL2/std.scad>;

	$fa = .01;
	$fs = $preview ? 5 : 1;
	$fn = 200;

	renderType = ""; // horzSlice, vertSlice, all
    
    screwHolesUp = 20;
    screwHoles = [5, 7.5, 10, 12.5, 15, 17.5, 20, 22.5, 25, 27.5];
    plateAttachHeight = 30;
    
    plateHeight = 30;

    plateType = "small"; // large small
    attachCyliderHeight = 16;
    plateBottomSize = 8;
    
    shiftTopX =3;
    shiftTopY = 3;
    
    hasGroove = false;
    grooveWidth=8;
    grooveDepth=2;
    
    grooveRounding=1;

	module hand_press_plate(plateWidth=30, plateDepth=30, plateAttachHeight=29, attachCyliderRadius=18, attachCyliderHeight=25, attachCyliderZOffset=4, screwHoleDiameter=2, holderScrewDimpleDepth=3, plateType=plateType, shiftTopX=15,shiftTopY=5){
		
        difference(){
        union(){
        //down(plateHeight+attachCyliderZOffset)
            cyl(l=attachCyliderHeight, d=attachCyliderRadius, rounding=2);
            rotate([180,0,0])
                prismoid(size1=[plateWidth,plateDepth], size2=[plateBottomSize,plateBottomSize], h=plateHeight, rounding=1, shift=[shiftTopX,shiftTopY]);
           
        }
        
        if (hasGroove) {
            move([shiftTopX,shiftTopY, -plateHeight])
            #cuboid([grooveWidth,40, grooveDepth], rounding=grooveRounding);

        }
        /*for (pos = screwHoles) {
            rotate([90, -90, 0])
            translate([pos-screwHolesUp-attachCyliderZOffset-plateAttachHeight, 0, attachCyliderRadius/2 ])  // adjust as needed
            
            #cyl(l=holderScrewDimpleDepth, d=screwHoleDiameter, rounding=1);
        }*/
        }
        
	}
    
    
    sliced(renderType=renderType){
        hand_press_plate(plateType=plateType,  attachCyliderHeight=attachCyliderHeight, plateAttachHeight=plateAttachHeight, shiftTopY=shiftTopY, shiftTopX =shiftTopX);
    }

     
module sliced(
    renderType = "horzSlice",        // "horzSlice", "vertSlice", or "all"
    sliceSize = 1000,
    sliceThickness = 0.3,
    showRawSlices = false,
    horzSlicePos = [-500, -500, 0],
    vertSlicePos = [0, -500, -500]
) {
   
    module horz_slice(raw=false) {
        if (raw) {
            translate(horzSlicePos)
                cube([sliceSize, sliceSize, sliceThickness], center=false);
        } else {
            intersection() {
                children();
                translate(horzSlicePos)
                    cube([sliceSize, sliceSize, sliceThickness], center=false);
            }
        }
    }

    module vert_slice(raw=false) {
        if (raw) {
            translate(vertSlicePos)
                cube([sliceThickness, sliceSize, sliceSize], center=false);
        } else {
            intersection() {
                children();
                translate(vertSlicePos)
                    cube([sliceThickness, sliceSize, sliceSize], center=false);
            }
        }
    }

    if (renderType == "horzSlice") {
        horz_slice(raw=showRawSlices){
            children();
        }
    } else if (renderType == "vertSlice") {
        vert_slice(raw=showRawSlices){
            children();
        }
    } else if (renderType == "all") {
        // show raw slices for reference
        horz_slice(raw=true);
        vert_slice(raw=true);
        // show full object
        children();
    } else {
        // show full object
        children();
    }
}
