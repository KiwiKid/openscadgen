

	include <BOSL2/std.scad>;
    include <BOSL2/joiners.scad>;
	$fa = .01;
	$fs = $preview ? 5 : 1;
	$fn = 200;
    
    function sum_col(matrix, col, i = 0) = 
    (i < len(matrix)) ? matrix[i][col] + sum_col(matrix, col, i + 1) : 0;

	/*
	renderType:
	use to print a test slice and confirm sizing before printing:
	 - "horzSlice" - horizontal slices (default)
	 - "vertSlice" - vertical slices
	 - "all" - the whole object
	*/
	renderType = "obj";
    
    
    





wallSize = 2;
sideWallSize = 4;
tinyAmount = 0.01;

buttonSquareHoleHeight = 33;
buttonSquareHoleWidth = 150;
buttonHoleRadius = 16;//16.3;
buttonSquareHole = [52,buttonSquareHoleWidth,buttonSquareHoleHeight];

// Hole Config
rowCount = 8;
colCount = 5;


/* Configure "set" of rows and columns here, for each entry in the main list [A, B, C, D, E], we will create a new set of button holes:
//          A: rowsConfiguredIndex number of rows of buttons holes in this set
            B: colsConfiguredIndex number of columns of button holes in this set
            C: rowOffsetIndex number of columns of button holes in this set
            D: arcHoleSpreadIndex 
            E: zSpacingIndex
*/
rowsConfiguredIndex = 0;
colsConfiguredIndex = 1;
rowOffsetIndex =2;
arcHoleSpreadIndex = 3;
zSpacingIndex = 4;
                  
holeConfig = [
 [rowCount, colCount, 0, 5.25, buttonSquareHoleHeight+wallSize],
 [1, 2, rowCount*buttonSquareHoleHeight*0.52+buttonSquareHoleHeight/2+5, 2.5, 0]
];

rowCountTotal = sum_col(holeConfig, 0);

maxColumnCount = max([for (row = holeConfig) row[1]]);

echo(maxColumnCount); // Output: ECHO: 7.5

boxWidth = 58*maxColumnCount; //240; //maxColumnCount* //240;

boxDepth = 63;

boxHeight = (rowCountTotal)*buttonSquareHoleHeight+30;


sectionHeight = boxHeight/2+20;
    topOffset  = sectionHeight/2;
    bottomOffset = -sectionHeight/2;
     section="topCorner"; //bottom all
     
    topCornerSectionMove= [0,100,boxHeight/2];
    topCornerSectionSize =  [150,150,150];
     
     
// Geometry math to align with the yscale(0.1) cylinder face
cylinder_radius = boxWidth*4;
holeOffset = 1.5;
holeUp = 33;
    
    holeAdjustVert = 5;
   holeAdjustHorz = 0;
   holeRotate = 18;
    
    arcAngleAmount = 5;
    
    faceCylScaling = 0.15;
    holderWedgeSize = [boxWidth,30, 10];
    holderBarSize = [boxWidth,30, 3];
    floorMode = "floor";
    
    screwRadius = 2;
    screwChamfer = 2;
    screwHeadRadius = screwRadius+screwChamfer;
    screwHoleLength = 8;
    screwHoleMove = [boxWidth/2,15,0];
    screwHoleOffset = boxWidth/3;
    
    module screw_hole(){
        cyl(r=screwRadius, h=screwHoleLength, chamfer2=-screwChamfer);
        up(screwHoleLength)
                cyl(r=screwHeadRadius, h=screwHoleLength);
        }
    
cutoutSize = [boxDepth+2, boxWidth-wallSize*2, boxHeight-wallSize*2];
    module main_box(){
        difference(){
        cuboid([boxDepth, boxWidth, boxHeight], $fn=64)
        children(); 
        right(tinyAmount)
        cuboid(cutoutSize, $fn=64);
        if(floorMode == "none"){
            down(10)
            cuboid(cutoutSize, $fn=64);
        }
        }
        
        rotate([0,0,90])
        move([-boxWidth/2, -boxDepth/2, boxHeight/2])
        rotate([-90,0,0])
        difference(){
        wedge(holderWedgeSize);
        move(screwHoleMove)
        left(screwHoleOffset)
        screw_hole();
        
        move(screwHoleMove)
        left(-screwHoleOffset)
        screw_hole();
        
        
        move(screwHoleMove)
        screw_hole();
        }
   /*     back(holderBarSize[1]/2)
        down(holderWedgeSize[2]/2-holderBarSize[2]/2)
        #cuboid(holderBarSize);
        */
        
    }
    alignmentHoleRadius = 3;
    alignmentHoleOffset = 20;
    alignmentHoleMove = [0,0,buttonSquareHoleHeight/2-wallSize*2];
    
   
    
    module button_hole(){
        move([holeAdjustVert,0,holeAdjustHorz])
        yrot(-holeRotate)
        ycyl(r=buttonHoleRadius, h=1000, $fn=5);
        back(245/2+2)
        cuboid(buttonSquareHole);
        left(alignmentHoleOffset)
        move(alignmentHoleMove)
        ycyl(r=alignmentHoleRadius, h=100);

        left(-alignmentHoleOffset)
        move(alignmentHoleMove)
        ycyl(r=alignmentHoleRadius, h=100);
    }
    
    
module holes(rows = 5, cols = 4, arcHoleSpread = 5, z_spacing=34){
    // Grid parameters matching the cabinet's face geometry
                // Number of items vertically (Z-axis)
               // Number of items per arc
    
   // Arc constraint to keep holes inside the usable front panel area
    start_angle = -arcHoleSpread;   
    end_angle = arcHoleSpread;   

    // Center the grid vertically on the cabinet face
    z_start = -((rows - 1) * z_spacing) / 2;

    for (i = [0 : rows - 1]) {
        z = z_start + (i * z_spacing); 
        
        for (j = [0 : cols - 1]) {
            // Distribute angles across the arc span
            angle = start_angle + (j / (cols - 1)) * (end_angle - start_angle);
            
            // 1. Move to vertical row position
            // 2. Rotate around Z to match the cylinder curve
            // 3. Move out to the cylinder's edge
            // 4. Compress the depth step by 0.1 to match the yscale(0.1) deformation
            translate([0, -cylinder_radius+holeOffset, z+boxHeight/2])
                rotate([0, 0, angle])
                    translate([0, cylinder_radius, 0])
                        scale([1, faceCylScaling, 1]) // Counteract the face scaling so holes stay circular
                            button_hole();
        }
    }
}


// rowCount, columnCount, rowVerticalOffset, buttonHorzontalGap, z_spacing


module dovetail_connector(){
dovetailWidthOffset = 60;
dovetailWidth = boxDepth+50-dovetailWidthOffset;
dovetailHeight = boxHeight-5;
dovetailMoveFwd = 2;

rotate([90,0,0])
    diff("remove")
   left(dovetailMoveFwd)
  cuboid([dovetailWidth+dovetailWidthOffset,sectionHeight,boxWidth]){
    attach(BACK) 
    dovetail("male", slide=boxWidth, width=dovetailWidth, height=18, chamfer=1);
    
    tag("remove")
    attach(FRONT) 
    dovetail("female", slide=boxWidth, width=dovetailWidth, height=18,chamfer=1);
  }

}

	module rounded_button_box(){
		main_box() {
    
    // Attach a rounded face to the TOP of the cuboid
    attach(LEFT, BOTTOM) 
        xrot(90)
        
        down(boxHeight/2){
        difference(){
        yscale(faceCylScaling)
            cyl(d=boxWidth, h=boxHeight);
            
            down(holeUp)
                       
                for(data = holeConfig){
                    rowsConfigured = data[rowsConfiguredIndex];
                    colsConfigured = data[colsConfiguredIndex];
                    rowOffset = data[rowOffsetIndex];
                    arcHoleSpread = data[arcHoleSpreadIndex];
                    z_spacing = data[zSpacingIndex];
                    
                    up(rowOffset)
                   holes(rows=rowsConfigured, cols=colsConfigured, arcHoleSpread=arcHoleSpread, z_spacing=z_spacing);

                }
                
            }
        }
        
        
    // Attach a rounded face to the FRONT of the cuboid (Optional Example)
    // attach(FRONT, BOTTOM) 
    //     cyl(d=40, l=20, rounding1=0, rounding2=20);
}
	}
    
    
    sliced(renderType=renderType) {
        if(section == "all"){
            rounded_button_box();
        } else if(section == "top"){
            intersection(){
                rounded_button_box();
                
                up(topOffset)
                 dovetail_connector();
            }
        } else if(section == "bottom"){
            intersection(){
                rounded_button_box();
                
                up(bottomOffset)
                 dovetail_connector();
            }
        } else if(section == "singleBox"){
            intersection(){
                rounded_button_box();
                
                fwd(28)
                down(1)
                 #cuboid([130,55,35]);
            }
        } else if(section == "topCorner"){
            intersection(){
                rounded_button_box();
                
                move(topCornerSectionMove)
                 #cuboid(topCornerSectionSize);
            }
        } else {
            echo("SECTION IS NOT VALID");
            }
    }
       








	
     
module sliced(
    renderType = "horzSlice",        // "horzSlice", "vertSlice", or "all"
    sliceSize = 1000,
    sliceThickness = 0.2,
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

