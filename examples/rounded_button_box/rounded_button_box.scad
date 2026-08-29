

	include <BOSL2/std.scad>;
    include <BOSL2/joiners.scad>;
    include <BOSL2/partitions.scad>;
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
sideWallSize = 3;
tinyAmount = 0.01;

// Hole the buttons sit in
buttonSquareHoleHeight = 33;
boxHeightExtra = 40;
buttonSquareHoleWidth = 150;
// Center button hole
buttonHoleRadius = 13.5;//16.3;
buttonSquareHole = [52,buttonSquareHoleWidth,buttonSquareHoleHeight];

// Hole Config
rowCount = 10;
colCount = 3;

zRowSpacingScaleConstant = 3.8;
zRowSpacingScale = 1.3;
boxWidthPerColumn = 65;

holeOffset = -1+colCount*colCount*0.5;//rowCount*1;
holeUp = 33;

/* Configure "set" of different rows and columns combinations here, for each entry in the main list [A, B, C, D, E], we will create a new set of button holes:
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
 [rowCount, colCount, 0, zRowSpacingScaleConstant+(colCount*zRowSpacingScale), buttonSquareHoleHeight+wallSize],
 [1, 2, rowCount*buttonSquareHoleHeight*0.52+buttonSquareHoleHeight/2+5, zRowSpacingScaleConstant+(1*zRowSpacingScale), 0]
];

rowCountTotal = sum_col(holeConfig, 0);

maxColumnCount = max([for (row = holeConfig) row[1]]);

echo(maxColumnCount); // Output: ECHO: 7.5

boxWidth = boxWidthPerColumn*maxColumnCount; //240; //maxColumnCount* //240;

boxDepth = 65;

boxHeight = (rowCountTotal)*buttonSquareHoleHeight+boxHeightExtra;


sectionHeight = boxHeight/2+20;


echo("boxHeight");
echo(boxHeight);
echo("boxWidth");
echo(boxWidth);

/// "holder-wall-mounted" 
    topOffset  = sectionHeight/2;
    bottomOffset = -sectionHeight/2;
     section="wallMount-withSpace-all"; //  "buttonJiggleBorder"; //bottom all devTopCorner  devTopSample devSingleBox cordPlugSide cordPlugBottom cordPlugNoHoleBottom holder 
     
     
     // wallMount-withSpace-withBox wallMount-withSpace-withWillBox
     
     
     // wallMount wallMount-withSpace wallMount-withSpace-bottom wallMount-withSpace-bottom
     /// wallMount-withSpace-withBox
     
// Geometry math to align with the yscale(0.1) cylinder face
cylinder_radius = boxWidth*2;
    
    holeAdjustVert = 0;
   holeAdjustHorz = 3;
   holeRotate = 0;
    
    arcAngleAmount = 5;
    
    faceCylScaling = 0.15;
    holderWedgeSize = [boxWidth,20, 20];
    holderBarSize = [boxWidth,30, 3];
    floorMode = "floor";
    
    screwRadius = 2;
    screwChamfer = 2;
    screwHeadRadius = screwRadius+screwChamfer;
    screwHoleLength = 8;
    screwHoleMove = [-boxHeight/2+70,0,-30];
    screwHoleOffset = boxWidth/3;
    
    module screw_hole(screwHoleLength=screwHoleLength){
        cyl(r=screwRadius, h=screwHoleLength, chamfer2=-screwChamfer);
        up(screwHoleLength)
                cyl(r=screwHeadRadius, h=screwHoleLength);
        }
    
cutoutSize = [boxDepth+2, boxWidth-sideWallSize*2, boxHeight-wallSize*2];
    module main_box(floorMode="none"){
        difference(){
        cuboid([boxDepth, boxWidth, boxHeight], $fn=64)
        children(); 
        right(tinyAmount)
        cuboid(cutoutSize, $fn=64);
        if(floorMode == "none"){
            down(10)
            cuboid(cutoutSize, $fn=64);
        }
        

        
        
        // cord gap hole
        /*move(holeLocation)
        xrot(90)
        yrot(90)
        cord_hole(mode="hole", plugDepth=wallSize);
        
        
        // cord gap hole                    
        move(holeLocationTwo)
        yrot(90)
        cord_hole(mode="hole", plugDepth=sideWallSize);*/
        }
        
        rotate([0,0,90])
        move([-boxWidth/2, -boxDepth/2, boxHeight/2-wallSize])
        rotate([-90,0,0])
        difference(){
        wedge(holderWedgeSize);
        /*move(screwHoleMove)
        left(screwHoleOffset)
        screw_hole();
        
        move(screwHoleMove)
        left(-screwHoleOffset)
        screw_hole();
        
        
        move(screwHoleMove)
        screw_hole();*/
        }
   /*     back(holderBarSize[1]/2)
        down(holderWedgeSize[2]/2-holderBarSize[2]/2)
        cuboid(holderBarSize);
        */
        
    }
    alignmentHoleRadius = 1.85;
    alignmentHoleOffset = 18.5;
    alignmentHoleVerticalOffset = 2.2;
    alignmentHoleMove = [0,0,buttonSquareHoleHeight/2-wallSize*2+alignmentHoleVerticalOffset];
    
   
    
    module button_hole(includeButton){
        move([holeAdjustVert,0,holeAdjustHorz])
        yrot(-holeRotate)
        ycyl(r=buttonHoleRadius, h=1000, $fn=8);
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


module screw_set(screwHoleLength=screwHoleLength){
    zcopies(50, 2){
                rotate([0,90,180]){
                move(screwHoleMove)
                fwd(screwHoleOffset)
                screw_hole(screwHoleLength=screwHoleLength);
                
                move(screwHoleMove)
                fwd(-screwHoleOffset)
                screw_hole(screwHoleLength=screwHoleLength);
                
                
                move(screwHoleMove)
                screw_hole(screwHoleLength=screwHoleLength); 
         }
        }
}

wallMountCutoutLipHeight = 12;

wallMountShiftUp = 100;
wallMountWidth = 55;
wallMountHeight = 120;
wallMountCubeSize = [cutoutSize[0]-wallMountWidth,cutoutSize[1],wallMountHeight];
wallMountCutoutSize = [(cutoutSize[0]-wallMountWidth)*0.9, cutoutSize[1]*0.98,wallMountHeight];
module wall_mount(){
    difference(){
        up(boxHeight/2-wallMountHeight/2-3)
        right(wallMountWidth/2-1.001)
        difference(){
        cuboid(wallMountCubeSize);
        left(5)
        down(wallMountCutoutLipHeight)
            cuboid(wallMountCutoutSize, rounding=5);
            

        }
        rotate([0,0,90])
                move([-boxWidth/2, -boxDepth/2, boxHeight/2-wallSize])
                rotate([-90,0,0])
               wedge(holderWedgeSize);
                
                
                screw_set();
        }
}

withSpaceExtraSpace = 80;
wallMountWithSpaceBackThickness = 1;

wallMountWithSpaceCubeSize = [withSpaceExtraSpace,boxWidth-0.001,boxHeight-0.001];

wallMountWithSpaceTopWidth = 25;

wallMountWithSpaceCutoutSize = [withSpaceExtraSpace, cutoutSize[1]*0.95,boxHeight-wallMountWithSpaceTopWidth-10];


//slideBottomCutoutMove =  [-30, 0, -boxHeight/2-5+0.05];
//slideBottomCutoutSize = [24, boxWidth+100, 24];
module wall_mount_with_space(cutOffset = 0, part = "all", wallMountExtraDepth=withSpaceExtraSpace) {
    // 1. Base Geometry
    screwSetRight = wallMountExtraDepth/2-10;
    cutoutRounding = 3;
    
    bottomHoleRadius = 15;
    bottomHoleLeft = -60;
    bottomHoleMove = [23, bottomHoleLeft, -boxHeight/2];
    
    module raw_part() {
        difference() {
            difference() {
                right(3+wallMountWithSpaceBackThickness)
                cuboid(wallMountWithSpaceCubeSize, rounding=0.3);
                down(wallMountWithSpaceTopWidth / 2 - 0.1)
                    cuboid(wallMountWithSpaceCutoutSize, chamfer=cutoutRounding);
                    
                   /* if(part == "all" || part == "bottom"){
                    move(slideBottomCutoutMove)
                    cuboid(slideBottomCutoutSize, anchor=BOT, chamfer=10, edges=RIGHT);
                    }*/
                    
                right(screwSetRight)
                if(part == "top" || part == "all"){
                
                    screw_set(screwHoleLength=wallMountWithSpaceCubeSize[1]-wallMountWithSpaceCutoutSize[1]);
                } else if(part == "bottom"){
                down(boxHeight/2)
                    screw_set(screwHoleLength=wallMountWithSpaceCubeSize[1]-wallMountWithSpaceCutoutSize[1]);
                    
                    
                    
                }
                
                if(part =="bottom" || part == "all"){
                    move(bottomHoleMove)
                        zcyl(r=bottomHoleRadius, h=20);
                    }
                
                
                    
            }
            left(wallMountExtraDepth / 2)
                rounded_button_box(floorMode = "on");
        /*        left(80)
                fwd(150)
                #cuboid([300,300,350]);*/
        }
    }

    // Dynamic bounding size for the cutting plane
    mask_size = max(wallMountWithSpaceCubeSize) * 3;

    // 2. Solid 3D Cut Mask
    module joiner_mask() {
        union() {
            // Main solid cutting block for the bottom volume
            down(mask_size / 2)
                cuboid([mask_size, mask_size, mask_size], anchor=CENTER);

            // 3D Male Dovetail Joiner along the seam
            right(20)
                
                xcopies(25, n=4) 
                dovetail(
                    "male", 
                    width = 14, 
                    height = 10, 
                    angle = 20, 
                    slide = mask_size,
                );
        }
    }

    // 3. Render Selection Logic
    if (part == "all") {
        raw_part();

    } else if (part == "top") {
        // Subtract mask (creates female pocket on top part)
        difference() {
            raw_part();
            up(cutOffset) 
                joiner_mask();
        }

    } else if (part == "bottom") {
        // Intersect mask (leaves male pin on bottom part)
        intersection() {
            raw_part();
            up(cutOffset) 
                joiner_mask();
        }
    }
}


module dovetail_connector(){
dovetailWidthOffset = 75;
dovetailWidth = boxDepth+50-dovetailWidthOffset;
dovetailHeight = 20;
dovetailMoveFwd = 2;

rotate([90,0,0])
    diff("remove")
   left(dovetailMoveFwd)
  cuboid([dovetailWidth+dovetailWidthOffset,sectionHeight,boxWidth]){
    attach(BACK) 
    dovetail("male", slide=boxWidth, width=dovetailWidth, height=dovetailHeight, chamfer=1);
    
    tag("remove")
    attach(FRONT) 
    dovetail("female", slide=boxWidth, width=dovetailWidth, height=dovetailHeight,chamfer=1);
  }

}

cordHoleSideOffset = 2;
cordHoleSize = 2.5;
cordHoleCornerOffset = 14;
cordInsertBarGapDown = 7;

holeSize = [30,sideWallSize+0.01,20];
holeRounding = 10;
holeLocation = [boxDepth/2, -boxWidth/2+cordHoleCornerOffset, -boxHeight/2+wallSize/2];

holeLocationTwo = [boxDepth/2, -boxWidth/2+sideWallSize/2, -boxHeight/2+wallSize/2+cordHoleCornerOffset];

module hole_shape(gender="male", plugDepth=2){
    dovetail(gender,h=10, w=12, slide=plugDepth,  radius=1);
    }

    
module cord_hole(mode="hole", plugDepth=2){
    coverWidthScale = 0.3;

    if(mode == "hole"){
        hole_shape(gender="female", plugDepth=plugDepth);
        //cuboid(holeSize, rounding=holeRounding, edges="Y");
    } else if(mode == "plug" || mode =="plugNoHoles"){

    
        difference(){
        union(){
        hole_shape(gender="male", plugDepth=plugDepth); 
        fwd(plugDepth*coverWidthScale)
        scale([1.2, coverWidthScale, 1.2])
        hole_shape(gender="male", plugDepth=plugDepth);     
      }
        if(mode != "plugNoHoles"){
            up(4)
            left(cordHoleSideOffset)
            ycyl(r=cordHoleSize, h=100);
            
            up(cordInsertBarGapDown)
            right(35)
            fwd(sideWallSize/2)
            cuboid([75,sideWallSize*3+0.02,2.7], rounding=1, edges="Y");
            }
        }
    
    }
    
    
}

	module rounded_button_box(holeConfig=holeConfig, floorMode=floorMode){
		main_box(floorMode=floorMode) {
    
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
        } else if(section == "devSingleBox"){
            intersection(){
                rounded_button_box();
                
//                down(buttonSquareHoleHeight/2+2)
                down(buttonSquareHoleHeight+3)
                left(58)
                 cuboid([30,55,buttonSquareHoleHeight+2]);
            }
        } else if(section == "buttonJiggleBorder"){
        
        buttonJiggleBorderSize = [46, 2, 29.5];
        
        buttonJiggleBorderCutoutSize = [45.2, 3, 28.5];
        
        difference(){
            cuboid(buttonJiggleBorderSize);
            cuboid(buttonJiggleBorderCutoutSize);
            }
        /*
        buttonSquareHoleHeight = 33;
boxHeightExtra = 40;
buttonSquareHoleWidth = 150;
// Center button hole
buttonHoleRadius = 13.5;//16.3;
buttonSquareHole = [52,buttonSquareHoleWidth,buttonSquareHoleHeight];
        
        
        
        borderScaleSize = 0.96;
        globalBorderWidthScale = 0.88;     globalBorderHeightScale = 0.87;

            scale([globalBorderWidthScale,1, globalBorderHeightScale])
            difference(){
                scale([1, 0.03,1])
                cuboid(buttonSquareHole);
                scale([borderScaleSize, 1.1,borderScaleSize])
                cuboid(buttonSquareHole);
            }
            */
        } else if(section == "devTopCorner"){
        
    topCornerSectionMove= [0,100,boxHeight/2];
    topCornerSectionSize =  [150,150,150];
     
            intersection(){
                rounded_button_box();
                
                move(topCornerSectionMove)
                 cuboid(topCornerSectionSize);
            }
        }else if(section == "devTopSample"){
        
    topCornerSectionMove= [0,100,boxHeight/2];
    topDevSectionSize =  [150,1000,100];
     
            intersection(){
                rounded_button_box();
                
                move(topCornerSectionMove)
                 cuboid(topDevSectionSize);
            }
        }else if(section == "cordPlugSide"){
            cord_hole(mode="plug", plugDepth=sideWallSize);
        }else if(section == "cordPlugBottom"){
            cord_hole(mode="plug", plugDepth=wallSize);
        }else if(section == "cordPlugNoHolesSide"){
            cord_hole(mode="plugNoHoles", plugDepth=sideWallSize);
        } else if(section == "cordPlugNoHolesBottom"){
            cord_hole(mode="plugNoHoles", plugDepth=wallSize);
        } else if(section == "wallMount"){
            wall_mount();
       } else if(section == "wallMount-withSpace-all" || section == "all") {      
            wall_mount_with_space(part="all"); //part="all" | "top" | "bottom"
        }else if(section == "wallMount-withSpace-withBox") {
        
                withBoxOffset = 10;
            
                left(withSpaceExtraSpace/2-withBoxOffset)
                rounded_button_box();
                right(withSpaceExtraSpace/2-withBoxOffset)
            wall_mount_with_space(part="all"); //part="all" | "top" | "bottom"
            
        }else if(section == "wallMount-withSpace-withWillBox") {
        
        
        
        
        
        
        
        
        
                withBoxOffset = 10;
            
                left(withSpaceExtraSpace/2-withBoxOffset)
                rounded_button_box();
                right(withSpaceExtraSpace/2-withBoxOffset)
            wall_mount_with_space(part="all", wallMountExtraDepth=300); //part="all" | "top" | "bottom"
            
        }else if(section == "wallMount-withSpace-top") {      
            wall_mount_with_space(part="top"); //part="all" | "top" | "bottom"
        }else if(section == "wallMount-withSpace-bottom") {      
            wall_mount_with_space(part="bottom"); //part="all" | "top" | "bottom"
        }else {
            echo("SECTION IS NOT VALID");
       }
    
       








	
     

